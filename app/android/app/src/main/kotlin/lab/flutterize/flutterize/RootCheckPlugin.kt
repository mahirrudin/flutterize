package lab.flutterize.flutterize

import android.content.pm.PackageManager
import android.os.Build
import io.flutter.plugin.common.MethodCall
import io.flutter.plugin.common.MethodChannel
import java.io.File

class RootCheckPlugin(private val activity: MainActivity) : MethodChannel.MethodCallHandler {

    override fun onMethodCall(call: MethodCall, result: MethodChannel.Result) {
        when (call.method) {
            "isRooted" -> result.success(isRooted())
            "isEmulator" -> result.success(isEmulator())
            "isCompromised" -> result.success(isRooted() || isEmulator())
            else -> result.notImplemented()
        }
    }

    private fun isRooted(): Boolean {
        return checkSuBinary() || checkRootBinaries() || checkRootPackages()
    }

    private fun checkSuBinary(): Boolean {
        val paths = arrayOf(
            "/system/bin/su",
            "/system/xbin/su",
            "/sbin/su",
            "/data/local/xbin/su",
            "/data/local/bin/su",
            "/system/sd/xbin/su",
            "/system/bin/failsafe/su",
            "/data/local/su"
        )
        return paths.any { File(it).exists() }
    }

    private fun checkRootBinaries(): Boolean {
        val binaries = arrayOf("busybox", "magisk", "supersu", "Superuser.apk")
        val paths = System.getenv("PATH")?.split(":") ?: return false
        return paths.any { path ->
            binaries.any { binary -> File(path, binary).exists() }
        }
    }

    private fun checkRootPackages(): Boolean {
        val packages = arrayOf(
            "com.topjohnwu.magisk",
            "eu.chainfire.supersu",
            "com.noshufou.android.su",
            "com.koushikdutta.superuser",
            "com.thirdparty.superuser",
            "com.yellowes.su"
        )
        val pm = activity.packageManager
        return packages.any {
            try {
                pm.getPackageInfo(it, 0)
                true
            } catch (e: PackageManager.NameNotFoundException) {
                false
            }
        }
    }

    private fun isEmulator(): Boolean {
        return (Build.FINGERPRINT.startsWith("google/sdk_gphone")
                || Build.FINGERPRINT.startsWith("generic")
                || Build.FINGERPRINT.startsWith("unknown")
                || Build.MODEL.contains("google_sdk")
                || Build.MODEL.contains("Emulator")
                || Build.MODEL.contains("Android SDK built for x86")
                || Build.MODEL.contains("sdk_gphone")
                || Build.BOARD == "goldfish"
                || Build.BOARD == "ranchu"
                || Build.HARDWARE.contains("goldfish")
                || Build.HARDWARE.contains("ranchu")
                || Build.PRODUCT.contains("sdk")
                || Build.PRODUCT.contains("emulator")
                || Build.MANUFACTURER.contains("Genymotion")
                || Build.BRAND.startsWith("generic") && Build.DEVICE.startsWith("generic"))
    }
}
