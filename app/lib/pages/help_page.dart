import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../services/auth_service.dart';

class HelpPage extends StatefulWidget {
  const HelpPage({super.key});

  @override
  State<HelpPage> createState() => _HelpPageState();
}

class _HelpPageState extends State<HelpPage> {
  bool _isLoading = true;

  // VULN #12: WebView-like page with JavaScript bridge
  // In a real WebView, this would expose getToken() to any loaded page
  // Here we simulate the concept with an InAppWebView-style widget

  @override
  void initState() {
    super.initState();
    _loadContent();
  }

  Future<void> _loadContent() async {
    // Simulate loading delay
    await Future.delayed(const Duration(milliseconds: 500));
    if (mounted) setState(() => _isLoading = false);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Help & FAQ'),
        actions: [
          // VULN: Debug button that exposes token — simulates JS bridge
          IconButton(
            icon: const Icon(Icons.bug_report),
            tooltip: 'Debug Info',
            onPressed: _showDebugInfo,
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator(color: Color(0xFFE94560)))
          : SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildFaqItem(
                    'How do I transfer points?',
                    'Go to your Profile → Points & Transfers → Send Points. Enter the recipient\'s email or phone number and the amount.',
                  ),
                  _buildFaqItem(
                    'How do I reset my password?',
                    'On the login screen, tap "Forgot Password?" and enter your email or phone. A reset token will be sent via email and SMS.',
                  ),
                  _buildFaqItem(
                    'Where can I find the reset token?',
                    'Check MockMail at http://localhost:8025 or MockSMS at http://localhost:7600 to view your reset token.',
                  ),
                  _buildFaqItem(
                    'Is my data secure?',
                    'All communication uses HTTPS with SSL pinning. Your password is hashed with bcrypt. JWT tokens expire after 24 hours.',
                  ),
                  _buildFaqItem(
                    'How do I contact support?',
                    'This is a cybersecurity learning application. Visit the project README for more information.',
                  ),
                  const SizedBox(height: 32),
                  // VULN #12: Token display area — simulates what a JS bridge would expose
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: const Color(0xFF0F3460),
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: const Color(0xFF0F3460)),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Debug Console',
                          style: TextStyle(
                            color: Colors.grey,
                            fontSize: 12,
                            fontFamily: 'monospace',
                          ),
                        ),
                        const SizedBox(height: 4),
                        const Text(
                          '// window.FlutterizeApp bridge active',
                          style: TextStyle(
                            color: Colors.green,
                            fontSize: 11,
                            fontFamily: 'monospace',
                          ),
                        ),
                        const Text(
                          '// Methods: getToken(), getUserData()',
                          style: TextStyle(
                            color: Colors.green,
                            fontSize: 11,
                            fontFamily: 'monospace',
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
    );
  }

  Widget _buildFaqItem(String question, String answer) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: ExpansionTile(
        title: Text(question, style: const TextStyle(fontWeight: FontWeight.w600)),
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
            child: Text(answer, style: const TextStyle(color: Colors.white70)),
          ),
        ],
      ),
    );
  }

  /// VULN #12: Simulates JS bridge — exposes JWT token and user data
  Future<void> _showDebugInfo() async {
    final token = await AuthService.getToken();
    if (!mounted) return;

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: const Color(0xFF16213E),
        title: const Text('JS Bridge Output', style: TextStyle(fontFamily: 'monospace', fontSize: 14)),
        content: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text('FlutterizeApp.getToken():', style: TextStyle(color: Colors.green, fontFamily: 'monospace', fontSize: 12)),
              const SizedBox(height: 4),
              SelectableText(
                token ?? 'null',
                style: const TextStyle(color: Colors.amber, fontFamily: 'monospace', fontSize: 11),
              ),
              const SizedBox(height: 16),
              ElevatedButton.icon(
                onPressed: () {
                  if (token != null) {
                    Clipboard.setData(ClipboardData(text: token));
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('Token copied to clipboard')),
                    );
                  }
                },
                icon: const Icon(Icons.copy, size: 16),
                label: const Text('Copy Token'),
                style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFF0F3460)),
              ),
            ],
          ),
        ),
        actions: [TextButton(onPressed: () => Navigator.pop(context), child: const Text('Close'))],
      ),
    );
  }
}
