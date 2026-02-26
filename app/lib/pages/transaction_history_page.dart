import 'package:flutter/material.dart';
import '../services/api_service.dart';
import '../services/auth_service.dart';

class TransactionHistoryPage extends StatefulWidget {
  const TransactionHistoryPage({super.key});

  @override
  State<TransactionHistoryPage> createState() => _TransactionHistoryPageState();
}

class _TransactionHistoryPageState extends State<TransactionHistoryPage> {
  List<dynamic>? _transactions;
  bool _isLoading = true;
  String? _error;
  Map<String, dynamic>? _profile;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
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

      final profile = await ApiService.getProfile(token);
      final transactions = await ApiService.getTransactionHistory(token);

      setState(() {
        _profile = profile;
        _transactions = transactions;
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
      appBar: AppBar(title: const Text('Transaction History')),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator(color: Color(0xFFE94560)))
          : _error != null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(_error!, style: const TextStyle(color: Colors.red)),
                      const SizedBox(height: 16),
                      ElevatedButton(onPressed: _loadData, child: const Text('Retry')),
                    ],
                  ),
                )
              : _transactions == null || _transactions!.isEmpty
                  ? const Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.receipt_long, size: 64, color: Colors.grey),
                          SizedBox(height: 16),
                          Text(
                            'No transactions yet',
                            style: TextStyle(color: Colors.grey, fontSize: 16),
                          ),
                        ],
                      ),
                    )
                  : RefreshIndicator(
                      onRefresh: _loadData,
                      color: const Color(0xFFE94560),
                      child: ListView.builder(
                        padding: const EdgeInsets.all(16),
                        itemCount: _transactions!.length,
                        itemBuilder: (context, index) {
                          final tx = _transactions![index];
                          final currentUserID = _profile?['id'];
                          final isSender = tx['sender_id'] == currentUserID;
                          final amount = tx['amount'] as int;

                          return Card(
                            margin: const EdgeInsets.only(bottom: 8),
                            child: ListTile(
                              leading: CircleAvatar(
                                backgroundColor: isSender
                                    ? Colors.red.withValues(alpha: 0.2)
                                    : Colors.green.withValues(alpha: 0.2),
                                child: Icon(
                                  isSender ? Icons.arrow_upward : Icons.arrow_downward,
                                  color: isSender ? Colors.red : Colors.green,
                                ),
                              ),
                              title: Text(
                                isSender
                                    ? 'Sent to ${tx['receiver_name']}'
                                    : 'Received from ${tx['sender_name']}',
                                style: const TextStyle(color: Colors.white),
                              ),
                              subtitle: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  if (tx['note'] != null && tx['note'].toString().isNotEmpty)
                                    Text(
                                      tx['note'],
                                      style: TextStyle(color: Colors.grey[500], fontSize: 12),
                                    ),
                                  Text(
                                    _formatDate(tx['created_at']),
                                    style: TextStyle(color: Colors.grey[600], fontSize: 11),
                                  ),
                                ],
                              ),
                              trailing: Text(
                                '${isSender ? '-' : '+'}$amount',
                                style: TextStyle(
                                  fontSize: 18,
                                  fontWeight: FontWeight.bold,
                                  color: isSender ? Colors.red : Colors.green,
                                ),
                              ),
                            ),
                          );
                        },
                      ),
                    ),
    );
  }

  String _formatDate(String? dateStr) {
    if (dateStr == null) return '';
    try {
      final date = DateTime.parse(dateStr);
      return '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')} '
          '${date.hour.toString().padLeft(2, '0')}:${date.minute.toString().padLeft(2, '0')}';
    } catch (_) {
      return dateStr;
    }
  }
}
