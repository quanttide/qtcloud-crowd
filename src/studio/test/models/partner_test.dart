import 'package:flutter_test/flutter_test.dart';
import 'package:qtcloud_crowd_studio/models/partner.dart';

void main() {
  group('Partner JSON 往返', () {
    test('完整字段序列化/反序列化一致', () {
      const partner = Partner(
        id: 'p1',
        name: '张三',
        type: PartnerType.agent,
        certified: true,
      );
      final decoded = Partner.fromJson(partner.toJson());
      expect(decoded.id, 'p1');
      expect(decoded.name, '张三');
      expect(decoded.type, PartnerType.agent);
      expect(decoded.certified, isTrue);
      expect(decoded.toJson(), partner.toJson());
    });

    test('certified 缺省为 false', () {
      final decoded = Partner.fromJson({'id': 'p2', 'name': '李四'});
      expect(decoded.certified, isFalse);
    });

    test('未知类型回退 agent', () {
      final decoded = Partner.fromJson({
        'id': 'p3',
        'name': '王五',
        'type': 'unknown',
      });
      expect(decoded.type, PartnerType.agent);
    });

    test('类型枚举全覆盖：channel/agent/training', () {
      expect(
        PartnerType.values,
        containsAll([PartnerType.channel, PartnerType.agent, PartnerType.training]),
      );
    });
  });

  group('Partner 认证', () {
    test('copyWith 可标记已认证', () {
      const partner = Partner(
        id: 'p1',
        name: '张三',
        type: PartnerType.channel,
        certified: false,
      );
      final certified = partner.copyWith(certified: true);
      expect(certified.certified, isTrue);
      expect(partner.certified, isFalse);
    });
  });
}
