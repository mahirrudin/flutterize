package lab.flutterize.flutterize

import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel

class MainActivity : FlutterActivity() {
    private val ROOT_CHANNEL = "lab.flutterize/root_check"

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, ROOT_CHANNEL)
            .setMethodCallHandler(RootCheckPlugin(this))
    }
}
