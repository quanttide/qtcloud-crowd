import '../models/task.dart';
import 'file_store.dart';

/// 任务仓储（DDD 仓储接口）。
abstract interface class TaskRepository {
  Future<List<Task>> findAll();
  Future<Task?> findById(String id);
  Future<void> save(Task task);
}

/// InMemory 实现（测试注入 / web 平台）。
class InMemoryTaskRepository implements TaskRepository {
  InMemoryTaskRepository([List<Task>? initial])
      : _items = {for (final t in initial ?? <Task>[]) t.id: t};

  final Map<String, Task> _items;

  @override
  Future<List<Task>> findAll() async => List.unmodifiable(_items.values);

  @override
  Future<Task?> findById(String id) async => _items[id];

  @override
  Future<void> save(Task task) async {
    _items[task.id] = task;
  }
}

/// LocalFile 实现：JSON 数组文件 + 原子写（web 平台不可用）。
class LocalFileTaskRepository implements TaskRepository {
  LocalFileTaskRepository(this.path);

  final String path;

  @override
  Future<List<Task>> findAll() async {
    final items = await readJsonList(path);
    return items.map(Task.fromJson).toList();
  }

  @override
  Future<Task?> findById(String id) async {
    final items = await readJsonList(path);
    for (final m in items) {
      if (m['id'] == id) return Task.fromJson(m);
    }
    return null;
  }

  @override
  Future<void> save(Task task) async {
    final items = await readJsonList(path);
    final idx = items.indexWhere((m) => m['id'] == task.id);
    if (idx >= 0) {
      items[idx] = task.toJson();
    } else {
      items.add(task.toJson());
    }
    await writeJsonList(path, items);
  }
}
