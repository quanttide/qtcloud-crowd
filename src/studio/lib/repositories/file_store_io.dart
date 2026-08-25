import 'dart:io';

/// dart:io 实现（桌面/移动端/测试）。
Future<String?> readFile(String path) async {
  final file = File(path);
  if (!await file.exists()) return null;
  return file.readAsString();
}

/// 原子写：先写 `$path.tmp` 再 rename，避免写一半损坏原文件。
Future<void> writeFileAtomic(String path, String content) async {
  final file = File(path);
  await file.parent.create(recursive: true);
  final tmp = File('$path.tmp');
  await tmp.writeAsString(content, flush: true);
  await tmp.rename(path);
}
