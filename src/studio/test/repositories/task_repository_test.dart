import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:qtcloud_crowd_studio/models/task.dart';
import 'package:qtcloud_crowd_studio/repositories/task_repository.dart';

Task sampleTask(String id, {TaskStatus status = TaskStatus.pending}) => Task(
      id: id,
      title: '任务$id',
      content: '内容$id',
      acceptanceCriteria: '准则$id',
      status: status,
    );

void main() {
  group('InMemoryTaskRepository', () {
    test('findAll 返回注入的初始数据', () async {
      final repo = InMemoryTaskRepository([sampleTask('t1'), sampleTask('t2')]);
      expect((await repo.findAll()).length, 2);
    });

    test('save 新增与覆盖（同 id 只保留一条）', () async {
      final repo = InMemoryTaskRepository();
      await repo.save(sampleTask('t1'));
      await repo.save(sampleTask('t1', status: TaskStatus.done));
      final all = await repo.findAll();
      expect(all.length, 1);
      expect(all.single.status, TaskStatus.done);
    });

    test('失败路径：findById 不存在返回 null', () async {
      final repo = InMemoryTaskRepository();
      expect(await repo.findById('missing'), isNull);
    });

    test('失败路径：空仓储 findAll 返回空列表', () async {
      final repo = InMemoryTaskRepository();
      expect(await repo.findAll(), isEmpty);
    });
  });

  group('LocalFileTaskRepository（原子写）', () {
    late Directory tempDir;
    late String path;

    setUp(() async {
      tempDir = await Directory.systemTemp.createTemp('task_repo_test');
      path = '${tempDir.path}/tasks.json';
    });

    tearDown(() async {
      await tempDir.delete(recursive: true);
    });

    test('失败路径：文件不存在 → findAll 空列表、findById null', () async {
      final repo = LocalFileTaskRepository(path);
      expect(await repo.findAll(), isEmpty);
      expect(await repo.findById('t1'), isNull);
    });

    test('save 后落盘，可重新读出（持久化）', () async {
      final repo = LocalFileTaskRepository(path);
      await repo.save(sampleTask('t1'));
      await repo.save(sampleTask('t2'));

      // 新实例重新读取（模拟重启）
      final repo2 = LocalFileTaskRepository(path);
      final all = await repo2.findAll();
      expect(all.length, 2);
      expect(all.first.title, '任务t1');
    });

    test('同 id 重复 save 覆盖而不是追加', () async {
      final repo = LocalFileTaskRepository(path);
      await repo.save(sampleTask('t1'));
      await repo.save(sampleTask('t1', status: TaskStatus.reviewing));
      expect((await LocalFileTaskRepository(path).findAll()).length, 1);
      expect(
        (await LocalFileTaskRepository(path).findById('t1'))!.status,
        TaskStatus.reviewing,
      );
    });

    test('原子写：写完后无 .tmp 残留文件', () async {
      final repo = LocalFileTaskRepository(path);
      await repo.save(sampleTask('t1'));
      expect(File('$path.tmp').existsSync(), isFalse);
      expect(File(path).existsSync(), isTrue);
    });
  });
}
