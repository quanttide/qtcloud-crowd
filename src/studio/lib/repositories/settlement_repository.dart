import '../models/settlement.dart';
import 'file_store.dart';

/// 结算仓储（DDD 仓储接口）。
abstract interface class SettlementRepository {
  Future<List<Settlement>> findAll();
  Future<Settlement?> findById(String id);
  Future<void> save(Settlement settlement);
}

/// InMemory 实现（测试注入 / web 平台）。
class InMemorySettlementRepository implements SettlementRepository {
  InMemorySettlementRepository([List<Settlement>? initial])
      : _items = {for (final s in initial ?? <Settlement>[]) s.id: s};

  final Map<String, Settlement> _items;

  @override
  Future<List<Settlement>> findAll() async => List.unmodifiable(_items.values);

  @override
  Future<Settlement?> findById(String id) async => _items[id];

  @override
  Future<void> save(Settlement settlement) async {
    _items[settlement.id] = settlement;
  }
}

/// LocalFile 实现：JSON 数组文件 + 原子写（web 平台不可用）。
class LocalFileSettlementRepository implements SettlementRepository {
  LocalFileSettlementRepository(this.path);

  final String path;

  @override
  Future<List<Settlement>> findAll() async {
    final items = await readJsonList(path);
    return items.map(Settlement.fromJson).toList();
  }

  @override
  Future<Settlement?> findById(String id) async {
    final items = await readJsonList(path);
    for (final m in items) {
      if (m['id'] == id) return Settlement.fromJson(m);
    }
    return null;
  }

  @override
  Future<void> save(Settlement settlement) async {
    final items = await readJsonList(path);
    final idx = items.indexWhere((m) => m['id'] == settlement.id);
    if (idx >= 0) {
      items[idx] = settlement.toJson();
    } else {
      items.add(settlement.toJson());
    }
    await writeJsonList(path, items);
  }
}
