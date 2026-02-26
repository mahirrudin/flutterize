import 'package:flutter/material.dart';
import '../services/api_service.dart';
import '../services/auth_service.dart';

class PointsPage extends StatefulWidget {
  const PointsPage({super.key});

  @override
  State<PointsPage> createState() => _PointsPageState();
}

class _PointsPageState extends State<PointsPage> {
  int? _balance;
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadBalance();
  }

  Future<void> _loadBalance() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final token = await AuthService.getToken();
      if (token == null) {
        if (mounted) Navigator.of(context).pushReplacementNamed('/login');
        return;
      }

      final response = await ApiService.getBalance(token);
      setState(() {
        _balance = response['balance'] as int?;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString().replaceAll('Exception: ', '');
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Points')),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator(color: Color(0xFFE94560)))
          : _error != null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(_error!, style: const TextStyle(color: Colors.red)),
                      const SizedBox(height: 16),
                      ElevatedButton(onPressed: _loadBalance, child: const Text('Retry')),
                    ],
                  ),
                )
              : RefreshIndicator(
                  onRefresh: _loadBalance,
                  color: const Color(0xFFE94560),
                  child: SingleChildScrollView(
                    physics: const AlwaysScrollableScrollPhysics(),
                    padding: const EdgeInsets.all(24),
                    child: Column(
                      children: [
                        Container(
                          width: double.infinity,
                          padding: const EdgeInsets.all(32),
                          decoration: BoxDecoration(
                            gradient: const LinearGradient(
                              colors: [Color(0xFF0F3460), Color(0xFF16213E)],
                              begin: Alignment.topLeft,
                              end: Alignment.bottomRight,
                            ),
                            borderRadius: BorderRadius.circular(16),
                            border: Border.all(color: const Color(0xFFE94560).withValues(alpha: 0.3)),
                          ),
                          child: Column(
                            children: [
                              const Icon(Icons.stars, size: 48, color: Color(0xFFE94560)),
                              const SizedBox(height: 12),
                              Text(
                                '${_balance ?? 0}',
                                style: const TextStyle(
                                  fontSize: 48,
                                  fontWeight: FontWeight.bold,
                                  color: Colors.white,
                                ),
                              ),
                              Text(
                                'Available Points',
                                style: TextStyle(fontSize: 14, color: Colors.grey[400]),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 32),
                        SizedBox(
                          width: double.infinity,
                          child: ElevatedButton.icon(
                            onPressed: () async {
                              await Navigator.of(context).pushNamed('/transfer-points');
                              _loadBalance(); // Refresh after transfer
                            },
                            icon: const Icon(Icons.send),
                            label: const Text('Send Points'),
                          ),
                        ),
                        const SizedBox(height: 12),
                        SizedBox(
                          width: double.infinity,
                          child: ElevatedButton.icon(
                            onPressed: () {
                              Navigator.of(context).pushNamed('/transaction-history');
                            },
                            icon: const Icon(Icons.history),
                            label: const Text('Transaction History'),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: const Color(0xFF0F3460),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
    );
  }
}
