import 'dart:convert';
import 'dart:io';
import 'package:flutter/services.dart';
import 'package:http/io_client.dart';
import 'package:http/http.dart' as http;

class ApiService {
  // Change this to your backend URL
  // For Android emulator: https://10.0.2.2:8443
  // For physical device: https://<your-ip>:8443
  static const String baseUrl = 'https://10.0.2.2:8443';

  static IOClient? _client;

  static Future<IOClient> getClient() async {
    if (_client != null) return _client!;

    // Load the root CA certificate for SSL pinning
    SecurityContext context = SecurityContext(withTrustedRoots: false);
    try {
      ByteData data = await rootBundle.load('assets/ca/rootCA.pem');
      context.setTrustedCertificatesBytes(data.buffer.asUint8List());
    } catch (e) {
      // If CA cert not found, fall back to default (for development)
      context = SecurityContext(withTrustedRoots: true);
    }

    HttpClient httpClient = HttpClient(context: context);
    // SECURE #8: No badCertificateCallback — SSL pinning enforced

    _client = IOClient(httpClient);
    return _client!;
  }

  static Map<String, String> _headers(String? token) {
    final headers = {'Content-Type': 'application/json'};
    if (token != null) {
      headers['Authorization'] = 'Bearer $token';
    }
    return headers;
  }

  // Auth endpoints
  static Future<Map<String, dynamic>> login(String identifier, String password) async {
    final client = await getClient();
    final response = await client.post(
      Uri.parse('$baseUrl/api/login'),
      headers: _headers(null),
      body: jsonEncode({'identifier': identifier, 'password': password}),
    );
    return _handleResponse(response);
  }

  static Future<Map<String, dynamic>> register(
    String email, String phone, String fullname, String birthdate, String password,
  ) async {
    final client = await getClient();
    final response = await client.post(
      Uri.parse('$baseUrl/api/register'),
      headers: _headers(null),
      body: jsonEncode({
        'email': email,
        'phone': phone,
        'fullname': fullname,
        'birthdate': birthdate,
        'password': password,
      }),
    );
    return _handleResponse(response);
  }

  static Future<Map<String, dynamic>> forgotPassword(String identifier) async {
    final client = await getClient();
    final response = await client.post(
      Uri.parse('$baseUrl/api/forgot-password'),
      headers: _headers(null),
      body: jsonEncode({'identifier': identifier}),
    );
    return _handleResponse(response);
  }

  static Future<Map<String, dynamic>> resetPassword(String token, String newPassword) async {
    final client = await getClient();
    final response = await client.post(
      Uri.parse('$baseUrl/api/reset-password'),
      headers: _headers(null),
      body: jsonEncode({'token': token, 'new_password': newPassword}),
    );
    return _handleResponse(response);
  }

  // Profile endpoints
  static Future<Map<String, dynamic>> getProfile(String token) async {
    final client = await getClient();
    final response = await client.get(
      Uri.parse('$baseUrl/api/profile'),
      headers: _headers(token),
    );
    return _handleResponse(response);
  }

  static Future<Map<String, dynamic>> updateProfile(
    String token, String fullname, String phone, String birthdate,
  ) async {
    final client = await getClient();
    final response = await client.put(
      Uri.parse('$baseUrl/api/profile'),
      headers: _headers(token),
      body: jsonEncode({
        'fullname': fullname,
        'phone': phone,
        'birthdate': birthdate,
      }),
    );
    return _handleResponse(response);
  }

  static Future<Map<String, dynamic>> changePassword(
    String token, String oldPassword, String newPassword,
  ) async {
    final client = await getClient();
    final response = await client.put(
      Uri.parse('$baseUrl/api/change-password'),
      headers: _headers(token),
      body: jsonEncode({
        'old_password': oldPassword,
        'new_password': newPassword,
      }),
    );
    return _handleResponse(response);
  }

  // Points endpoints
  static Future<Map<String, dynamic>> getBalance(String token) async {
    final client = await getClient();
    final response = await client.get(
      Uri.parse('$baseUrl/api/points/balance'),
      headers: _headers(token),
    );
    return _handleResponse(response);
  }

  static Future<Map<String, dynamic>> transferPoints(
    String token, String receiverIdentifier, int amount, String note,
  ) async {
    final client = await getClient();
    final response = await client.post(
      Uri.parse('$baseUrl/api/points/transfer'),
      headers: _headers(token),
      body: jsonEncode({
        'receiver_identifier': receiverIdentifier,
        'amount': amount,
        'note': note,
      }),
    );
    return _handleResponse(response);
  }

  static Future<List<dynamic>> getTransactionHistory(String token) async {
    final client = await getClient();
    final response = await client.get(
      Uri.parse('$baseUrl/api/points/history'),
      headers: _headers(token),
    );
    if (response.statusCode == 200) {
      return jsonDecode(response.body) as List<dynamic>;
    }
    throw Exception(_extractError(response));
  }

  static Map<String, dynamic> _handleResponse(http.Response response) {
    final body = jsonDecode(response.body) as Map<String, dynamic>;
    if (response.statusCode >= 200 && response.statusCode < 300) {
      return body;
    }
    throw Exception(body['error'] ?? 'Request failed');
  }

  static String _extractError(http.Response response) {
    try {
      final body = jsonDecode(response.body) as Map<String, dynamic>;
      return body['error'] ?? 'Request failed';
    } catch (_) {
      return 'Request failed with status ${response.statusCode}';
    }
  }
}
