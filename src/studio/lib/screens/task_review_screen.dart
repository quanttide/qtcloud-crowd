import 'package:flutter/material.dart';

import '../models/task.dart';
import '../repositories/task_repository.dart';

/// 事：任务审核——两个判断：能不能发（审核）+ 算不算过（验收）。
class TaskReviewScreen extends StatefulWidget {
  const TaskReviewScreen({super.key, required this.repository});

  final TaskRepository repository;

  @override
  State<TaskReviewScreen> createState() => _TaskReviewScreenState();
}

class _TaskReviewScreenState extends State<TaskReviewScreen> {
  List<Task> _tasks = [];
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final tasks = await widget.repository.findAll();
      setState(() {
        _tasks = tasks;
        _error = null;
      });
    } catch (e) {
      setState(() => _error = '加载任务失败：$e');
    }
  }

  /// 审核发布：pending → reviewing（验收准则清晰才可发布）。
  Future<void> _publish(Task task) async {
    if (!task.canPublish) return;
    await widget.repository.save(task.copyWith(status: TaskStatus.reviewing));
    await _load();
  }

  /// 验收通过：reviewing → done。
  Future<void> _approve(Task task) async {
    await widget.repository.save(task.copyWith(status: TaskStatus.done));
    await _load();
  }

  /// 打回：reviewing → pending（按准则未达标，退回补全）。
  Future<void> _reject(Task task) async {
    await widget.repository.save(task.copyWith(status: TaskStatus.pending));
    await _load();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('任务审核')),
      body: _error != null
          ? Center(child: Text(_error!))
          : ListView.builder(
              itemCount: _tasks.length,
              itemBuilder: (context, index) {
                final task = _tasks[index];
                return Card(
                  margin: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 6,
                  ),
                  child: Padding(
                    padding: const EdgeInsets.all(12),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Expanded(
                              child: Text(
                                task.title,
                                style: const TextStyle(
                                  fontWeight: FontWeight.bold,
                                  fontSize: 16,
                                ),
                              ),
                            ),
                            _StatusChip(status: task.status),
                          ],
                        ),
                        const SizedBox(height: 4),
                        Text('内容：${task.content}'),
                        Text('验收准则：${task.acceptanceCriteria}'),
                        const SizedBox(height: 8),
                        _buildActions(task),
                      ],
                    ),
                  ),
                );
              },
            ),
    );
  }

  Widget _buildActions(Task task) {
    switch (task.status) {
      case TaskStatus.pending:
        // 审核（能不能发）：验收准则说不清就不能发布。
        return Align(
          alignment: Alignment.centerRight,
          child: FilledButton.icon(
            onPressed: task.canPublish ? () => _publish(task) : null,
            icon: const Icon(Icons.send),
            label: Text(task.canPublish ? '审核发布' : '验收准则缺失，不能发布'),
          ),
        );
      case TaskStatus.reviewing:
        return Row(
          mainAxisAlignment: MainAxisAlignment.end,
          children: [
            OutlinedButton.icon(
              onPressed: () => _reject(task),
              icon: const Icon(Icons.replay),
              label: const Text('打回'),
            ),
            const SizedBox(width: 8),
            FilledButton.icon(
              onPressed: () => _approve(task),
              icon: const Icon(Icons.check),
              label: const Text('验收通过'),
            ),
          ],
        );
      case TaskStatus.done:
        return const Align(
          alignment: Alignment.centerRight,
          child: Text('已完成', style: TextStyle(color: Colors.green)),
        );
    }
  }
}

class _StatusChip extends StatelessWidget {
  const _StatusChip({required this.status});

  final TaskStatus status;

  @override
  Widget build(BuildContext context) {
    final (label, color) = switch (status) {
      TaskStatus.pending => ('待审核', Colors.orange),
      TaskStatus.reviewing => ('审核中', Colors.blue),
      TaskStatus.done => ('已完成', Colors.green),
    };
    return Chip(label: Text(label), backgroundColor: color.withValues(alpha: 0.15));
  }
}
