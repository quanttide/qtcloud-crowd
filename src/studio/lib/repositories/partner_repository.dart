import '../models/partner.dart';
import 'file_store.dart';

/// 执行方仓储（DDD 仓储接口）。
abstract interface class PartnerRepository {
  Future<List<Partner>> findAll();
  Future<Partner?> findById(String id);
  Future<void> save(Partner partner);
}

/// InMemory 实现（测试注入 / web 平台）。
class InMemoryPartnerRepository implements PartnerRepository {
  InMemoryPartnerRepository([List<Partner>? initial])
      : _items = {for (final p in initial ?? <Partner>[]) p.id: p};

  final Map<String, Partner> _items;

  @override
  Future<List<Partner>> findAll() async => List.unmodifiable(_items.values);

  @override
  Future<Partner?> findById(String id) async => _items[id];

  @override
  Future<void> save(Partner partner) async {
    _items[partner.id] = partner;
  }
}

/// LocalFile 实现：JSON 数组文件 + 原子写（web 平台不可用）。
class LocalFilePartnerRepository implements PartnerRepository {
  LocalFilePartnerRepository(this.path);

  final String path;

  @override
  Future<List<Partner>> findAll() async {
    final items = await readJsonList(path);
    return items.map(Partner.fromJson).toList();
  }

  @override
  Future<Partner?> findById(String id) async {
    final items = await readJsonList(path);
    for (final m in items) {
      if (m['id'] == id) return Partner.fromJson(m);
    }
    return null;
  }

  @override
  Future<void> save(Partner partner) async {
    final items = await readJsonList(path);
    final idx = items.indexWhere((m) => m['id'] == partner.id);
    if (idx >= 0) {
      items[idx] = partner.toJson();
    } else {
      items.add(partner.toJson());
    }
    await writeJsonList(path, items);
  }
}
