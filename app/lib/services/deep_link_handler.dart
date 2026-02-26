import 'dart:async';
import 'package:flutter/services.dart';
import 'api_service.dart';
import 'auth_service.dart';

class DeepLinkHandler {
  static const _channel = MethodChannel('lab.flutterize/deep_links');
  static StreamController<Uri>? _linkController;

  static Stream<Uri> get linkStream {
    _linkController ??= StreamController<Uri>.broadcast();
    return _linkController!.stream;
  }

  static void init() {
    _channel.setMethodCallHandler((call) async {
      if (call.method == 'onDeepLink') {
        final uri = Uri.parse(call.arguments as String);
        _linkController?.add(uri);
      }
    });
  }

  /// VULN #11: Handles deep link actions WITHOUT re-authentication
  static Future<Map<String, dynamic>?> handleUri(Uri uri) async {
    final token = await AuthService.getToken();
    if (token == null) return null;

    switch (uri.host) {
      case 'action':
        return _handleAction(uri, token);
      default:
        return null;
    }
  }

  static Future<Map<String, dynamic>?> _handleAction(Uri uri, String token) async {
    final path = uri.pathSegments;
    if (path.isEmpty) return null;

    switch (path[0]) {
      case 'transfer':
        final to = uri.queryParameters['to'];
        final amountStr = uri.queryParameters['amount'];
        final note = uri.queryParameters['note'] ?? 'Deep link transfer';
        if (to != null && amountStr != null) {
          final amount = int.tryParse(amountStr);
          if (amount != null) {
            // VULN: Executes transfer without user confirmation
            return await ApiService.transferPoints(token, to, amount, note);
          }
        }
        break;

      case 'reset':
        final resetToken = uri.queryParameters['token'];
        final newPassword = uri.queryParameters['new_password'];
        if (resetToken != null && newPassword != null) {
          // VULN: Resets password without user interaction
          return await ApiService.resetPassword(resetToken, newPassword);
        }
        break;
    }
    return null;
  }
}
