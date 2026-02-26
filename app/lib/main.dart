import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'pages/login_page.dart';
import 'pages/profile_page.dart';
import 'pages/change_password_page.dart';
import 'pages/forgot_password_page.dart';
import 'pages/register_page.dart';
import 'pages/reset_password_page.dart';
import 'pages/points_page.dart';
import 'pages/transfer_points_page.dart';
import 'pages/transaction_history_page.dart';
import 'pages/help_page.dart';
import 'services/auth_service.dart';
import 'services/root_detector.dart';

void main() {
  runApp(const FlutterizeApp());
}

class FlutterizeApp extends StatelessWidget {
  const FlutterizeApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Flutterize',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        brightness: Brightness.dark,
        primarySwatch: Colors.indigo,
        scaffoldBackgroundColor: const Color(0xFF1A1A2E),
        appBarTheme: const AppBarTheme(
          backgroundColor: Color(0xFF16213E),
          elevation: 0,
        ),
        inputDecorationTheme: InputDecorationTheme(
          filled: true,
          fillColor: const Color(0xFF16213E),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: const BorderSide(color: Color(0xFF0F3460)),
          ),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: const BorderSide(color: Color(0xFF0F3460)),
          ),
          focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: const BorderSide(color: Color(0xFFE94560), width: 2),
          ),
        ),
        elevatedButtonTheme: ElevatedButtonThemeData(
          style: ElevatedButton.styleFrom(
            backgroundColor: const Color(0xFFE94560),
            foregroundColor: Colors.white,
            padding: const EdgeInsets.symmetric(vertical: 16),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
            ),
          ),
        ),
        textButtonTheme: TextButtonThemeData(
          style: TextButton.styleFrom(
            foregroundColor: const Color(0xFFE94560),
          ),
        ),
        cardTheme: CardThemeData(
          color: const Color(0xFF16213E),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
            side: const BorderSide(color: Color(0xFF0F3460)),
          ),
        ),
      ),
      home: const AuthGate(),
      routes: {
        '/login': (context) => const LoginPage(),
        '/register': (context) => const RegisterPage(),
        '/profile': (context) => const ProfilePage(),
        '/change-password': (context) => const ChangePasswordPage(),
        '/forgot-password': (context) => const ForgotPasswordPage(),
        '/reset-password': (context) => const ResetPasswordPage(),
        '/points': (context) => const PointsPage(),
        '/transfer-points': (context) => const TransferPointsPage(),
        '/transaction-history': (context) => const TransactionHistoryPage(),
        '/help': (context) => const HelpPage(),
      },
    );
  }
}

class AuthGate extends StatefulWidget {
  const AuthGate({super.key});

  @override
  State<AuthGate> createState() => _AuthGateState();
}

class _AuthGateState extends State<AuthGate> {
  @override
  void initState() {
    super.initState();
    _checkSecurity();
  }

  Future<void> _checkSecurity() async {
    // Root/emulator detection (always active)
    final isCompromised = await RootDetector.isDeviceCompromised();
    if (isCompromised && mounted) {
      _showSecurityAlert();
      return;
    }
    _checkAuth();
  }

  Future<void> _checkAuth() async {
    final isLoggedIn = await AuthService.isLoggedIn();
    if (mounted) {
      Navigator.of(context).pushReplacementNamed(
        isLoggedIn ? '/profile' : '/login',
      );
    }
  }

  void _showSecurityAlert() {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        backgroundColor: const Color(0xFF16213E),
        title: const Row(
          children: [
            Icon(Icons.warning_amber_rounded, color: Color(0xFFE94560), size: 28),
            SizedBox(width: 8),
            Text('Security Alert', style: TextStyle(color: Color(0xFFE94560))),
          ],
        ),
        content: const Text(
          'This device appears to be rooted or running on an emulator.\n\n'
          'For security reasons, this app cannot run on compromised devices.',
          style: TextStyle(color: Colors.white70),
        ),
        actions: [
          ElevatedButton(
            onPressed: () => _exitApp(),
            style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFFE94560)),
            child: const Text('Exit'),
          ),
        ],
      ),
    );
  }

  void _exitApp() {
    // This will close the app
    WidgetsBinding.instance.addPostFrameCallback((_) {
      SystemNavigator.pop();
    });
  }

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: CircularProgressIndicator(color: Color(0xFFE94560)),
      ),
    );
  }
}
