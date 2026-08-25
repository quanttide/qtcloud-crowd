import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:qtcloud_crowd_studio/models/partner.dart';
import 'package:qtcloud_crowd_studio/models/settlement.dart';
import 'package:qtcloud_crowd_studio/models/task.dart';
import 'package:qtcloud_crowd_studio/repositories/partner_repository.dart';
import 'package:qtcloud_crowd_studio/repositories/settlement_repository.dart';
import 'package:qtcloud_crowd_studio/repositories/task_repository.dart';
import 'package:qtcloud_crowd_studio/screens/partner_screen.dart';
import 'package:qtcloud_crowd_studio/screens/settlement_screen.dart';
import 'package:qtcloud_crowd_studio/screens/task_review_screen.dart';

Widget wrap(Widget child) => MaterialApp(home: child);

void main() {
  group('TaskReviewScreen', () {
    testWidgets('渲染任务列表：标题、状态、按钮', (tester) async {
      final repo = InMemoryTaskRepository([
        Task(
          id: 't1',
          title: '撰写财报摘要',
          content: '阅读财报',
          acceptanceCriteria: '摘要覆盖收入利润现金流',
          status: TaskStatus.pending,
        ),
        Task(
          id: 't2',
          title: '校对文档',
          content: '校对翻译',
          acceptanceCriteria: '无错别字',
          status: TaskStatus.reviewing,
        ),
        Task(
          id: 't3',
          title: '已完成任务',
          content: 'X',
          acceptanceCriteria: 'Y',
          status: TaskStatus.done,
        ),
      ]);
      await tester.pumpWidget(wrap(TaskReviewScreen(repository: repo)));
      await tester.pumpAndSettle();

      expect(find.text('撰写财报摘要'), findsOneWidget);
      expect(find.text('待审核'), findsOneWidget);
      expect(find.text('审核发布'), findsOneWidget);
      expect(find.text('验收通过'), findsOneWidget);
      expect(find.text('打回'), findsOneWidget);
      expect(find.text('已完成'), findsNWidgets(2)); // chip + done 标签
    });

    testWidgets('验收准则缺失 → 发布按钮禁用并提示', (tester) async {
      final repo = InMemoryTaskRepository([
        Task(
          id: 't1',
          title: '说不清验收的任务',
          content: '做点事',
          acceptanceCriteria: '',
          status: TaskStatus.pending,
        ),
      ]);
      await tester.pumpWidget(wrap(TaskReviewScreen(repository: repo)));
      await tester.pumpAndSettle();

      expect(find.text('验收准则缺失，不能发布'), findsOneWidget);
      final button = tester.widget<FilledButton>(
        find.widgetWithText(FilledButton, '验收准则缺失，不能发布'),
      );
      expect(button.onPressed, isNull);
    });

    testWidgets('点击验收通过 → 状态变为已完成', (tester) async {
      final repo = InMemoryTaskRepository([
        Task(
          id: 't1',
          title: '任务A',
          content: 'X',
          acceptanceCriteria: 'Y',
          status: TaskStatus.reviewing,
        ),
      ]);
      await tester.pumpWidget(wrap(TaskReviewScreen(repository: repo)));
      await tester.pumpAndSettle();

      await tester.tap(find.text('验收通过'));
      await tester.pumpAndSettle();

      final updated = await repo.findById('t1');
      expect(updated!.status, TaskStatus.done);
      // 屏幕刷新为已完成
      expect(find.text('已完成'), findsWidgets);
    });
  });

  group('PartnerScreen', () {
    testWidgets('渲染名单：姓名、类型、认证状态、按钮', (tester) async {
      final repo = InMemoryPartnerRepository([
        const Partner(
          id: 'p1',
          name: '张三',
          type: PartnerType.channel,
          certified: false,
        ),
        const Partner(
          id: 'p2',
          name: '李四',
          type: PartnerType.training,
          certified: true,
        ),
      ]);
      await tester.pumpWidget(wrap(PartnerScreen(repository: repo)));
      await tester.pumpAndSettle();

      expect(find.text('张三'), findsOneWidget);
      expect(find.text('李四'), findsOneWidget);
      expect(find.textContaining('渠道'), findsOneWidget);
      expect(find.textContaining('实训基地'), findsOneWidget);
      expect(find.textContaining('待认证'), findsOneWidget);
      expect(find.textContaining('已认证'), findsOneWidget);
      expect(find.text('认证'), findsOneWidget);
    });

    testWidgets('点击认证 → certified 变 true', (tester) async {
      final repo = InMemoryPartnerRepository([
        const Partner(
          id: 'p1',
          name: '张三',
          type: PartnerType.agent,
          certified: false,
        ),
      ]);
      await tester.pumpWidget(wrap(PartnerScreen(repository: repo)));
      await tester.pumpAndSettle();

      await tester.tap(find.text('认证'));
      await tester.pumpAndSettle();

      expect((await repo.findById('p1'))!.certified, isTrue);
    });
  });

  group('SettlementScreen', () {
    testWidgets('渲染结算记录列表与新增表单', (tester) async {
      final repo = InMemorySettlementRepository([
        const Settlement(
          id: 's1',
          taskId: 't1',
          partnerId: 'p1',
          amount: 888.5,
          settledAt: '2026-08-25T10:00:00.000Z',
        ),
      ]);
      await tester.pumpWidget(wrap(SettlementScreen(repository: repo)));
      await tester.pumpAndSettle();

      expect(find.text('新增一笔'), findsOneWidget);
      expect(find.text('任务 ID'), findsOneWidget);
      expect(find.text('执行方 ID'), findsOneWidget);
      expect(find.text('¥ 888.50'), findsOneWidget);
      expect(find.textContaining('t1'), findsWidgets);
      expect(find.text('记一笔'), findsOneWidget);
    });

    testWidgets('记一笔：金额合法 → 记录新增', (tester) async {
      final repo = InMemorySettlementRepository();
      await tester.pumpWidget(wrap(SettlementScreen(repository: repo)));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField).at(0), 't9');
      await tester.enterText(find.byType(TextField).at(1), 'p9');
      await tester.enterText(find.byType(TextField).at(2), '66');
      await tester.tap(find.text('记一笔'));
      await tester.pumpAndSettle();

      final all = await repo.findAll();
      expect(all.length, 1);
      expect(all.single.taskId, 't9');
      expect(all.single.amount, 66);
      expect(find.text('¥ 66.00'), findsOneWidget);
    });

    testWidgets('失败路径：金额不合法 → 提示且不新增', (tester) async {
      final repo = InMemorySettlementRepository();
      await tester.pumpWidget(wrap(SettlementScreen(repository: repo)));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField).at(0), 't9');
      await tester.enterText(find.byType(TextField).at(1), 'p9');
      await tester.enterText(find.byType(TextField).at(2), '-5');
      await tester.tap(find.text('记一笔'));
      await tester.pumpAndSettle();

      expect(find.text('请填写任务、执行方和大于 0 的金额'), findsOneWidget);
      expect(await repo.findAll(), isEmpty);
    });
  });
}
