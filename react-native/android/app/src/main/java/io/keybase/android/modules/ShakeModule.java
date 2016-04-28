package io.keybase.android.modules;

import android.content.Context;
import android.hardware.SensorManager;
import android.util.Log;

import com.facebook.react.bridge.ReactApplicationContext;
import com.facebook.react.bridge.ReactContextBaseJavaModule;
import com.facebook.react.common.ShakeDetector;
import com.facebook.react.modules.core.DeviceEventManagerModule;

import java.util.HashMap;
import java.util.Map;

public class ShakeModule extends ReactContextBaseJavaModule implements KillableModule {
    private static final String NAME = "ShakeModule";
    private static final String SHAKE_EVENT_NAME = "SHAKE";
    private final ShakeDetector shakeDetector;
    private boolean isShakeDetectorStarted = false;

    public ShakeModule(final ReactApplicationContext reactContext) {
        super(reactContext);

        shakeDetector = new ShakeDetector(new ShakeDetector.ShakeListener() {
            @Override
            public void onShake() {
                if (!reactContext.hasActiveCatalystInstance()) {
                    Log.w(NAME, "JS Bridge is dead, dropping shake message");
                }

                reactContext
                  .getJSModule(DeviceEventManagerModule.RCTDeviceEventEmitter.class)
                  .emit(SHAKE_EVENT_NAME, null);
            }
        });

        if (!isShakeDetectorStarted) {
            shakeDetector.start(
              (SensorManager) reactContext.getSystemService(Context.SENSOR_SERVICE));
            isShakeDetectorStarted = true;
        }
    }

    @Override
    public Map<String, Object> getConstants() {
        final Map<String, Object> constants = new HashMap<>();
        constants.put("eventName", SHAKE_EVENT_NAME);
        return constants;
    }

    @Override
    public void destroy() {
        if (isShakeDetectorStarted) {
            shakeDetector.stop();
            isShakeDetectorStarted = false;
        }
    }

    @Override
    public String getName() {
        return NAME;
    }
}
