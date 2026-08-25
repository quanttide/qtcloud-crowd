import 'package:flutter_test/flutter_test.dart';
import 'package:qtcloud_crowd_studio/models/task.dart';

void main() {
  group('Task JSON 往返', () {
    test('完整字段序列化/反序列化一致', () {
      const task = Task(
        id: 't1',
        title: '撰写季度财报摘要',
        content: '阅读财报并输出 3 点摘要',
        acceptanceCriteria: '摘要覆盖收入、利润、现金流',
        status: TaskStatus.pending,
      );
      final decoded = Task.fromJson(task.toJson());
      expect(decoded.id, 't1');
      expect(decoded.title, '撰写季度财报摘要');
      expect(decoded.content, '阅读财报并输出 3 点摘要');
      expect(decoded.acceptanceCriteria, '摘要覆盖收入、利润、现金流');
      expect(decoded.status, TaskStatus.pending);
      expect(decoded.toJson(), task.toJson());
    });

    test('缺失字段有兜底（content/acceptance_criteria 缺省为空串）', () {
      final decoded = Task.fromJson({'id': 't2', 'title': 'X'});
      expect(decoded.content, '');
      expect(decoded.acceptanceCriteria, '');
      expect(decoded.status, TaskStatus.pending);
    });

    test('未知状态枚举回退 pending', () {
      final decoded = Task.fromJson({
        'id': 't3',
        'title': 'X',
        'status': 'unknown_status',
      });
      expect(decoded.status, TaskStatus.pending);
    });
  });

  group('Task 验收准则兜底（模型层约束）', () {
    Task taskWith(String criteria) => Task(
          id: 't',
          title: 'X',
          content: 'Y',
          acceptanceCriteria: criteria,
          status: TaskStatus.pending,
        );

    test('验收准则为空 → 不能发布', () {
      expect(taskWith('').canPublish, isFalse);
    });

    test('验收准则全空白 → 不能发布', () {
      expect(taskWith('   \n  ').canPublish, isFalse);
    });

    test('验收准则非空 → 可以发布', () {
      expect(taskWith('交付文档通过评审').canPublish, isTrue);
    });
  });

  group('Task 状态流', () {
    test('pending → reviewing → done', () {
      Task t = Task(
        id: 't',
        title: 'X',
        content: 'Y',
        acceptanceCriteria: 'Z',
        status: TaskStatus.pending,
      );
      t = t.copyWith(status: TaskStatus.reviewing);
      expect(t.status, TaskStatus.reviewing);
      t = t.copyWith(status: TaskStatus.done);
      expect(t.status, TaskStatus.done);
    });
  });
}
