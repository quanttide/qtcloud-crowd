import 'package:flutter/material.dart';

import '../models/partner.dart';
import '../repositories/partner_repository.dart';

/// 人：执行方管理——名单 + 认证状态，一个动作：认证。
class PartnerScreen extends StatefulWidget {
  const PartnerScreen({super.key, required this.repository});

  final PartnerRepository repository;

  @override
  State<PartnerScreen> createState() => _PartnerScreenState();
}

class _PartnerScreenState extends State<PartnerScreen> {
  List<Partner> _partners = [];
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final partners = await widget.repository.findAll();
      setState(() {
        _partners = partners;
        _error = null;
      });
    } catch (e) {
      setState(() => _error = '加载执行方失败：$e');
    }
  }

  Future<void> _certify(Partner partner) async {
    await widget.repository.save(partner.copyWith(certified: true));
    await _load();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('执行方')),
      body: _error != null
          ? Center(child: Text(_error!))
          : ListView.builder(
              itemCount: _partners.length,
              itemBuilder: (context, index) {
                final partner = _partners[index];
                final typeLabel = switch (partner.type) {
                  PartnerType.channel => '渠道',
                  PartnerType.agent => '代理',
                  PartnerType.training => '实训基地',
                };
                return Card(
                  margin: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 6,
                  ),
                  child: ListTile(
                    leading: CircleAvatar(
                      child: Text(partner.name.characters.first),
                    ),
                    title: Text(partner.name),
                    subtitle: Text('$typeLabel · ${partner.certified ? "已认证" : "待认证"}'),
                    trailing: partner.certified
                        ? const Icon(Icons.verified, color: Colors.green)
                        : OutlinedButton(
                            onPressed: () => _certify(partner),
                            child: const Text('认证'),
                          ),
                  ),
                );
              },
            ),
    );
  }
}
