import 'package:flutter_test/flutter_test.dart';
import 'package:qtcloud_crowd_studio/models/settlement.dart';

void main() {
  group('Settlement JSON 往返', () {
    test('完整字段序列化/反序列化一致', () {
      const settlement = Settlement(
        id: 's1',
        taskId: 't1',
        partnerId: 'p1',
        amount: 888.5,
        settledAt: '2026-08-25T10:00:00.000Z',
      );
      final decoded = Settlement.fromJson(settlement.toJson());
      expect(decoded.id, 's1');
      expect(decoded.taskId, 't1');
      expect(decoded.partnerId, 'p1');
      expect(decoded.amount, 888.5);
      expect(decoded.settledAt, '2026-08-25T10:00:00.000Z');
      expect(decoded.toJson(), settlement.toJson());
    });

    test('amount 整数值也可反序列化（num → double）', () {
      final decoded = Settlement.fromJson({
        'id': 's2',
        'task_id': 't2',
        'partner_id': 'p2',
        'amount': 100,
        'settled_at': '2026-08-25T11:00:00.000Z',
      });
      expect(decoded.amount, 100.0);
    });
  });
}
