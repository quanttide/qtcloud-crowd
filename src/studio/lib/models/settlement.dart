/// 结算模型（财：结算管理）。
///
/// 结算的本质是"验收通过 → 记录付款"——钱从哪来、付给谁、付了多少。
library;

class Settlement {
  const Settlement({
    required this.id,
    required this.taskId,
    required this.partnerId,
    required this.amount,
    required this.settledAt,
  });

  final String id;
  final String taskId;
  final String partnerId;

  /// 结算金额（元）。
  final double amount;

  /// 结算时间（ISO 8601）。
  final String settledAt;

  factory Settlement.fromJson(Map<String, dynamic> json) {
    return Settlement(
      id: json['id'] as String,
      taskId: json['task_id'] as String,
      partnerId: json['partner_id'] as String,
      amount: (json['amount'] as num).toDouble(),
      settledAt: json['settled_at'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'task_id': taskId,
      'partner_id': partnerId,
      'amount': amount,
      'settled_at': settledAt,
    };
  }
}
