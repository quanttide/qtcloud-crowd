import 'package:flutter/material.dart';

import '../models/settlement.dart';
import '../repositories/settlement_repository.dart';

/// 财：结算管理——验收通过后记一笔（任务 + 执行方 + 金额 + 时间）。
class SettlementScreen extends StatefulWidget {
  const SettlementScreen({super.key, required this.repository});

  final SettlementRepository repository;

  @override
  State<SettlementScreen> createState() => _SettlementScreenState();
}

class _SettlementScreenState extends State<SettlementScreen> {
  List<Settlement> _settlements = [];
  String? _error;

  final _taskIdController = TextEditingController();
  final _partnerIdController = TextEditingController();
  final _amountController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _taskIdController.dispose();
    _partnerIdController.dispose();
    _amountController.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    try {
      final settlements = await widget.repository.findAll();
      setState(() {
        _settlements = settlements;
        _error = null;
      });
    } catch (e) {
      setState(() => _error = '加载结算记录失败：$e');
    }
  }

  /// 记一笔：金额必须 > 0，任务/执行方必须填写。
  Future<void> _addSettlement() async {
    final taskId = _taskIdController.text.trim();
    final partnerId = _partnerIdController.text.trim();
    final amount = double.tryParse(_amountController.text.trim());
    if (taskId.isEmpty || partnerId.isEmpty || amount == null || amount <= 0) {
      setState(() => _error = '请填写任务、执行方和大于 0 的金额');
      return;
    }
    final settlement = Settlement(
      id: 's_${DateTime.now().millisecondsSinceEpoch}',
      taskId: taskId,
      partnerId: partnerId,
      amount: amount,
      settledAt: DateTime.now().toIso8601String(),
    );
    await widget.repository.save(settlement);
    _taskIdController.clear();
    _partnerIdController.clear();
    _amountController.clear();
    await _load();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('结算记录')),
      body: ListView(
        padding: const EdgeInsets.all(12),
        children: [
          Card(
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('新增一笔', style: TextStyle(fontWeight: FontWeight.bold)),
                  const SizedBox(height: 8),
                  TextField(
                    controller: _taskIdController,
                    decoration: const InputDecoration(
                      labelText: '任务 ID',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    controller: _partnerIdController,
                    decoration: const InputDecoration(
                      labelText: '执行方 ID',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    controller: _amountController,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(
                      labelText: '金额（元）',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 8),
                  Align(
                    alignment: Alignment.centerRight,
                    child: FilledButton.icon(
                      onPressed: _addSettlement,
                      icon: const Icon(Icons.add),
                      label: const Text('记一笔'),
                    ),
                  ),
                ],
              ),
            ),
          ),
          if (_error != null)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 8),
              child: Text(_error!, style: const TextStyle(color: Colors.red)),
            ),
          const Divider(),
          for (final s in _settlements)
            Card(
              margin: const EdgeInsets.symmetric(vertical: 4),
              child: ListTile(
                leading: const Icon(Icons.payments),
                title: Text('¥ ${s.amount.toStringAsFixed(2)}'),
                subtitle: Text('任务 ${s.taskId} · 执行方 ${s.partnerId}'),
                trailing: Text(
                  s.settledAt.split('T').first,
                  style: const TextStyle(color: Colors.grey),
                ),
              ),
            ),
        ],
      ),
    );
  }
}
