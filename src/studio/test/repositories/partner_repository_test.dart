import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:qtcloud_crowd_studio/models/partner.dart';
import 'package:qtcloud_crowd_studio/repositories/partner_repository.dart';

void main() {
  group('InMemoryPartnerRepository', () {
    test('findAll 返回注入数据；save 覆盖更新', () async {
      final repo = InMemoryPartnerRepository();
      const p = Partner(
        id: 'p1',
        name: '张三',
        type: PartnerType.agent,
        certified: false,
      );
      await repo.save(p);
      await repo.save(p.copyWith(certified: true));
      expect((await repo.findAll()).length, 1);
      expect((await repo.findById('p1'))!.certified, isTrue);
    });

    test('失败路径：findById 不存在返回 null、findAll 空列表', () async {
      final repo = InMemoryPartnerRepository();
      expect(await repo.findById('missing'), isNull);
      expect(await repo.findAll(), isEmpty);
    });
  });

  group('LocalFilePartnerRepository', () {
    late Directory tempDir;
    late String path;

    setUp(() async {
      tempDir = await Directory.systemTemp.createTemp('partner_repo_test');
      path = '${tempDir.path}/partners.json';
    });

    tearDown(() async {
      await tempDir.delete(recursive: true);
    });

    test('失败路径：文件不存在 → 空列表', () async {
      expect(await LocalFilePartnerRepository(path).findAll(), isEmpty);
    });

    test('认证状态变更可持久化', () async {
      final repo = LocalFilePartnerRepository(path);
      const p = Partner(
        id: 'p1',
        name: '李四',
        type: PartnerType.channel,
        certified: false,
      );
      await repo.save(p);
      await repo.save(p.copyWith(certified: true));
      final reloaded = await LocalFilePartnerRepository(path).findById('p1');
      expect(reloaded!.certified, isTrue);
    });
  });
}
