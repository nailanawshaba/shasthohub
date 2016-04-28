import NativeEventEmitter from './native-event-emitter'
import {NativeModules} from 'react-native'

const shakeModule = NativeModules.ShakeModule

export default cb => NativeEventEmitter.addListener(shakeModule.eventName, cb)
