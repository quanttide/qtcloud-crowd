/// web 平台没有 dart:io——LocalFile 仓储不可用，运行时报错。
Future<String?> readFile(String path) {
  throw UnsupportedError('LocalFile 仓储不支持 web 平台，请改用 InMemory');
}

Future<void> writeFileAtomic(String path, String content) {
  throw UnsupportedError('LocalFile 仓储不支持 web 平台，请改用 InMemory');
}
