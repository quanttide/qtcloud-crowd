/// 执行方模型（人：执行方管理）。
///
/// 认证是唯一"不说清就不能往下走"的动作——接单前必须先知道对方是谁。
library;

enum PartnerType {
  /// 渠道
  channel,

  /// 代理
  agent,

  /// 实训基地成员
  training;

  static PartnerType fromJson(String? value) {
    return PartnerType.values.firstWhere(
      (t) => t.name == value,
      orElse: () => PartnerType.agent,
    );
  }
}

class Partner {
  const Partner({
    required this.id,
    required this.name,
    required this.type,
    required this.certified,
  });

  final String id;
  final String name;
  final PartnerType type;

  /// 认证状态：true = 已认证（可信执行方）。
  final bool certified;

  Partner copyWith({bool? certified}) {
    return Partner(
      id: id,
      name: name,
      type: type,
      certified: certified ?? this.certified,
    );
  }

  factory Partner.fromJson(Map<String, dynamic> json) {
    return Partner(
      id: json['id'] as String,
      name: json['name'] as String,
      type: PartnerType.fromJson(json['type'] as String?),
      certified: json['certified'] as bool? ?? false,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'type': type.name,
      'certified': certified,
    };
  }
}
