import 'package:flutter/services.dart';

class RootDetector {
  static const _channel = MethodChannel('lab.flutterize/root_check');

  static Future<bool> isRooted() async {
    try {
      return await _channel.invokeMethod('isRooted') ?? false;
    } catch (e) {
      return false;
    }
  }

  static Future<bool> isEmulator() async {
    try {
      return await _channel.invokeMethod('isEmulator') ?? false;
    } catch (e) {
      return false;
    }
  }

  static Future<bool> isDeviceCompromised() async {
    try {
      return await _channel.invokeMethod('isCompromised') ?? false;
    } catch (e) {
      return false;
    }
  }
}
