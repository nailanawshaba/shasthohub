// @flow
import {AppState} from 'react-native'

function appInBackground() : boolean {
    return AppState.currentState == 'background'
}

export {
  appInBackground,
}