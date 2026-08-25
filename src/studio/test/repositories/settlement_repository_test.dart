import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:qtcloud_crowd_studio/models/settlement.dart';
import 'package:qtcloud_crowd_studio/repositories/settlement_repository.dart';

void main() {
  group('InMemorySettlementRepository', () {
    test('save 后 findById 可读；覆盖不重复', () async {
      final repo = InMemorySettlementRepository();
      const s = Settlement(
        id: 's1',
        taskId: 't1',
        partnerId: 'p1',
        amount: 100,
        settledAt: '2026-08-25T10:00:00.000Z',
      );
      await repo.save(s);
      await repo.save(s);
      expect((await repo.findAll()).length, 1);
      expect((await repo.findById('s1'))!.amount, 100);
    });

    test('失败路径：findById 不存在返回 null、findAll 空列表', () async {
      final repo = InMemorySettlementRepository();
      expect(await repo.findById('missing'), isNull);
      expect(await repo.findAll(), isEmpty);
    });
  });

  group('LocalFileSettlementRepository', () {
    late Directory tempDir;
    late String path;

    setUp(() async {
      tempDir = await Directory.systemTemp.createTemp('settlement_repo_test');
      path = '${tempDir.path}/settlements.json';
    });

    tearDown(() async {
      await tempDir.delete(recursive: true);
    });

    test('失败路径：文件不存在 → 空列表', () async {
      expect(await LocalFileSettlementRepository(path).findAll(), isEmpty);
    });

    test('记一笔后可持久化读出', () async {
      final repo = LocalFileSettlementRepository(path);
      await repo.save(const Settlement(
        id: 's1',
        taskId: 't1',
        partnerId: 'p1',
        amount: 888.5,
        settledAt: '2026-08-25T10:00:00.000Z',
      ));
      final reloaded = await LocalFileSettlementRepository(path).findAll();
      expect(reloaded.single.amount, 888.5);
      expect(reloaded.single.taskId, 't1');
    });
  });
}
