import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import 'repositories/partner_repository.dart';
import 'repositories/settlement_repository.dart';
import 'repositories/task_repository.dart';
import 'screens/partner_screen.dart';
import 'screens/settlement_screen.dart';
import 'screens/task_review_screen.dart';

/// 仓储集合（三件套接线入口）。
class AppRepositories {
  const AppRepositories({
    required this.tasks,
    required this.partners,
    required this.settlements,
  });

  final TaskRepository tasks;
  final PartnerRepository partners;
  final SettlementRepository settlements;
}

/// 非 web 平台用 LocalFile（JSON 原子写持久化）；web 平台无 dart:io 用 InMemory。
AppRepositories createRepositories() {
  if (kIsWeb) {
    return AppRepositories(
      tasks: InMemoryTaskRepository(),
      partners: InMemoryPartnerRepository(),
      settlements: InMemorySettlementRepository(),
    );
  }
  return AppRepositories(
    tasks: LocalFileTaskRepository('data/tasks.json'),
    partners: LocalFilePartnerRepository('data/partners.json'),
    settlements: LocalFileSettlementRepository('data/settlements.json'),
  );
}

void main() {
  runApp(CrowdStudioApp(repositories: createRepositories()));
}

/// 量潮众包管理云——管理方后台，人（执行方）/财（结算）/事（任务）三视角。
class CrowdStudioApp extends StatelessWidget {
  const CrowdStudioApp({super.key, required this.repositories});

  final AppRepositories repositories;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: '量潮众包管理云',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.teal),
        useMaterial3: true,
      ),
      home: HomeShell(repositories: repositories),
    );
  }
}

class HomeShell extends StatefulWidget {
  const HomeShell({super.key, required this.repositories});

  final AppRepositories repositories;

  @override
  State<HomeShell> createState() => _HomeShellState();
}

class _HomeShellState extends State<HomeShell> {
  int _index = 0;

  @override
  Widget build(BuildContext context) {
    final pages = [
      TaskReviewScreen(repository: widget.repositories.tasks),
      PartnerScreen(repository: widget.repositories.partners),
      SettlementScreen(repository: widget.repositories.settlements),
    ];
    return Scaffold(
      body: pages[_index],
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (i) => setState(() => _index = i),
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.task_alt),
            label: '事·任务',
          ),
          NavigationDestination(
            icon: Icon(Icons.people),
            label: '人·执行方',
          ),
          NavigationDestination(
            icon: Icon(Icons.payments),
            label: '财·结算',
          ),
        ],
      ),
    );
  }
}
