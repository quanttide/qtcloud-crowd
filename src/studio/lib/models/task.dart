/// 任务模型（事：任务审核）。
///
/// 生命周期：pending（待审核）→ reviewing（审核中）→ done（已完成）。
/// 审核本质是两个判断：
/// 1. 能不能发（验收准则清晰吗）——发布前必须校验 [acceptanceCriteria] 非空；
/// 2. 算不算过（交付达标吗）——按准则判通过或打回。
library;

enum TaskStatus {
  /// 待审核（待发布）
  pending,

  /// 审核中（已发布）
  reviewing,

  /// 已完成（验收通过）
  done;

  static TaskStatus fromJson(String? value) {
    return TaskStatus.values.firstWhere(
      (s) => s.name == value,
      orElse: () => TaskStatus.pending,
    );
  }
}

class Task {
  const Task({
    required this.id,
    required this.title,
    required this.content,
    required this.acceptanceCriteria,
    required this.status,
  });

  final String id;
  final String title;

  /// 内容清单（任务要做什么）
  final String content;

  /// 验收准则（怎么算交付达标）
  final String acceptanceCriteria;

  final TaskStatus status;

  /// 验收准则兜底：说不清验收 = 不能发布（模型层约束）。
  bool get canPublish => acceptanceCriteria.trim().isNotEmpty;

  Task copyWith({
    String? id,
    String? title,
    String? content,
    String? acceptanceCriteria,
    TaskStatus? status,
  }) {
    return Task(
      id: id ?? this.id,
      title: title ?? this.title,
      content: content ?? this.content,
      acceptanceCriteria: acceptanceCriteria ?? this.acceptanceCriteria,
      status: status ?? this.status,
    );
  }

  factory Task.fromJson(Map<String, dynamic> json) {
    return Task(
      id: json['id'] as String,
      title: json['title'] as String,
      content: json['content'] as String? ?? '',
      acceptanceCriteria: json['acceptance_criteria'] as String? ?? '',
      status: TaskStatus.fromJson(json['status'] as String?),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'content': content,
      'acceptance_criteria': acceptanceCriteria,
      'status': status.name,
    };
  }
}
