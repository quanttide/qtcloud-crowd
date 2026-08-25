/// 原子 JSON 文件存储。
///
/// - 读：文件不存在/为空 → 返回空列表；
/// - 写：先写临时文件再 rename（原子写，避免半写坏文件）。
///
/// web 平台没有 dart:io，LocalFile 仓储不可用——由
/// `file_store_stub.dart` 提供运行时报错，main.dart 在 web 下改用 InMemory。
library;

import 'dart:convert';

import 'file_store_stub.dart'
    if (dart.library.io) 'file_store_io.dart' as store;

/// 读取 JSON 列表文件；文件不存在或内容为空时返回空列表。
Future<List<Map<String, dynamic>>> readJsonList(String path) async {
  final raw = await store.readFile(path);
  if (raw == null || raw.trim().isEmpty) return [];
  final decoded = jsonDecode(raw);
  if (decoded is! List) return [];
  return decoded.whereType<Map<String, dynamic>>().toList();
}

/// 原子写入 JSON 列表文件。
Future<void> writeJsonList(
  String path,
  List<Map<String, dynamic>> items,
) {
  final content = const JsonEncoder.withIndent('  ').convert(items);
  return store.writeFileAtomic(path, content);
}
