import 'package:flutter/material.dart';
import '../services/api_service.dart';
import '../services/auth_service.dart';

class ProfilePage extends StatefulWidget {
  const ProfilePage({super.key});

  @override
  State<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends State<ProfilePage> {
  Map<String, dynamic>? _profile;
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadProfile();
  }

  Future<void> _loadProfile() async {
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
      setState(() {
        _profile = profile;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString().replaceAll('Exception: ', '');
        _isLoading = false;
      });
    }
  }

  Future<void> _logout() async {
    await AuthService.clearToken();
    if (mounted) {
      Navigator.of(context).pushReplacementNamed('/login');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Profile'),
        automaticallyImplyLeading: false,
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: _logout,
            tooltip: 'Logout',
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator(color: Color(0xFFE94560)))
          : _error != null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(_error!, style: const TextStyle(color: Colors.red)),
                      const SizedBox(height: 16),
                      ElevatedButton(
                        onPressed: _loadProfile,
                        child: const Text('Retry'),
                      ),
                    ],
                  ),
                )
              : RefreshIndicator(
                  onRefresh: _loadProfile,
                  color: const Color(0xFFE94560),
                  child: SingleChildScrollView(
                    physics: const AlwaysScrollableScrollPhysics(),
                    padding: const EdgeInsets.all(24),
                    child: Column(
                      children: [
                        const CircleAvatar(
                          radius: 50,
                          backgroundColor: Color(0xFF0F3460),
                          child: Icon(Icons.person, size: 50, color: Color(0xFFE94560)),
                        ),
                        const SizedBox(height: 16),
                        Text(
                          _profile?['fullname'] ?? '',
                          style: const TextStyle(
                            fontSize: 24,
                            fontWeight: FontWeight.bold,
                            color: Colors.white,
                          ),
                        ),
                        const SizedBox(height: 24),
                        _buildInfoCard(Icons.email_outlined, 'Email', _profile?['email'] ?? ''),
                        _buildInfoCard(Icons.phone_outlined, 'Phone', _profile?['phone'] ?? ''),
                        _buildInfoCard(Icons.cake_outlined, 'Birthdate', _formatDate(_profile?['birthdate'] ?? '')),
                        _buildInfoCard(
                          Icons.stars_outlined,
                          'Points',
                          '${_profile?['points_balance'] ?? 0}',
                        ),
                        const SizedBox(height: 32),
                        SizedBox(
                          width: double.infinity,
                          child: ElevatedButton.icon(
                            onPressed: () {
                              Navigator.of(context).pushNamed('/points');
                            },
                            icon: const Icon(Icons.stars),
                            label: const Text('Points & Transfers'),
                          ),
                        ),
                        const SizedBox(height: 12),
                        SizedBox(
                          width: double.infinity,
                          child: ElevatedButton.icon(
                            onPressed: () async {
                              await Navigator.of(context).pushNamed('/change-password');
                            },
                            icon: const Icon(Icons.lock_outline),
                            label: const Text('Change Password'),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: const Color(0xFF0F3460),
                            ),
                          ),
                        ),
                        const SizedBox(height: 12),
                        SizedBox(
                          width: double.infinity,
                          child: ElevatedButton.icon(
                            onPressed: () {
                              Navigator.of(context).pushNamed('/help');
                            },
                            icon: const Icon(Icons.help_outline),
                            label: const Text('Help & FAQ'),
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

  Widget _buildInfoCard(IconData icon, String label, String value) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: ListTile(
        leading: Icon(icon, color: const Color(0xFFE94560)),
        title: Text(label, style: TextStyle(fontSize: 12, color: Colors.grey[500])),
        subtitle: Text(
          value,
          style: const TextStyle(fontSize: 16, color: Colors.white),
        ),
      ),
    );
  }

  String _formatDate(String dateStr) {
    if (dateStr.isEmpty) return '';
    // Take only YYYY-MM-DD, strip any time/timezone suffix
    if (dateStr.length >= 10) return dateStr.substring(0, 10);
    return dateStr;
  }
}
