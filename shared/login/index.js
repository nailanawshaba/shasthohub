// @flow
import ErrorText from './error.render'
import Intro from './forms/intro'
import Login from './login'
import React from 'react'
import {connect} from 'react-redux'
import signupRouter from './signup'
import {Map} from 'immutable'
import * as actions from '../actions/login/creators'
import HiddenString, {throwIfNotHidden} from '../util/hidden-string'

import Passphrase from './register/passphrase'
import UsernameOrEmail from './register/username-or-email'
import ErrorInLogin from './register/error'
import SelectOtherDevice from './register/select-other-device'
import GPGSign from './register/gpg-sign'
import PaperKey from './register/paper-key'
import CodePage from './register/code-page'
import SetPublicName from './register/set-public-name'
import SuccessRender from './signup/success/index'

import type {URI} from '../constants/router'
import type {TypedState} from '../constants/reducer'

function elementForPath (currentPath: Map<string, string>, uri: URI) {
  const [props, path] = (([currentPath.get('props'), currentPath.get('path')]): any)
  switch (path) {
    case 'codePage':
      return (
        <CodePage />
      )
    case 'errorInLogin':
      return (
        <DispatchHOC>
          {({dispatch}) => (
            <ErrorInLogin
              error={props.error}
              onBack={() => dispatch(actions.onBack())} />
          )}
        </DispatchHOC>
      )
    case 'chooseGPGMethod':
      return (
        <DispatchHOC>
          {({dispatch}) => (
            <GPGSign
              onSubmit={exportKey => dispatch(actions.chooseGPGMethod(exportKey))}
              onBack={() => dispatch(actions.onBack())} />
          )}
        </DispatchHOC>
      )
    case 'usernameOrEmail':
      return (
        <DispatchHOC>
          {({dispatch}) => (
            <UsernameOrEmail
              onSubmit={usernameOrEmail => dispatch(actions.startLogin(usernameOrEmail))}
              onBack={() => dispatch(actions.onBack())} />
          )}
        </DispatchHOC>
      )
    case 'passphrase':
      return (
        <DispatchHOC>
          {({dispatch}) => (
            <Passphrase
              prompt={props.prompt}
              onSubmit={passphrase => dispatch(actions.submitPassphrase(
                new HiddenString(passphrase),
                false,
              ))}
              onBack={() => dispatch(actions.onBack())}
              error={props.retryLabel}
              username={props.username} />
          )}
        </DispatchHOC>
      )
    case 'paperkey':
      return (
        <DispatchHOC>
          {({dispatch}) => (
            <PaperKey
              onSubmit={passphrase => dispatch(actions.submitPassphrase(
                new HiddenString(passphrase),
                false,
              ))}
              onBack={() => dispatch(actions.onBack())}
              error={props.retryLabel} />
          )}
        </DispatchHOC>
      )
    case 'chooseDevice':
      return (
        <DispatchHOC>
          {({dispatch}) => (
            <SelectOtherDevice
              devices={props.devices}
              onSelect={deviceId => dispatch(actions.selectDeviceId(deviceId))}
              onWont={() => dispatch(actions.onWont())}
              onBack={() => dispatch(actions.onBack())} />
          )}
        </DispatchHOC>
      )
    case 'promptDeviceName':
      return (
        <DispatchHOC>
          {({dispatch}) => (
            <SetPublicName
              existingDevices={props.existingDevices}
              deviceNameError={props.errorMessage}
              onSubmit={deviceName => dispatch(actions.submitDeviceName(deviceName))}
              onBack={() => dispatch(actions.onBack())} />
          )}
        </DispatchHOC>
      )
    case 'displayPaperKey':
      return (
        <DispatchHOC>
          {({dispatch}) => (
            <SuccessRender
              paperkey={throwIfNotHidden(props.paperkey)}
              waiting={false}
              onFinish={() => dispatch(actions.onFinish())}
              onBack={() => dispatch(actions.onBack())} />
          )}
        </DispatchHOC>
      )
    case 'root':
      return <Intro />
    case 'login':
      return <Login />
  }
}

function loginRouter (currentPath: Map<string, string>, uri: URI): any {
  // Fallback (for debugging)
  let element = <ErrorText currentPath={currentPath} />

  const parseRoute: any = currentPath.get('parseRoute')
  let {componentAtTop: {component: Component, props, element: dynamicElement}} = parseRoute || {componentAtTop: {}}

  if (dynamicElement) {
    element = dynamicElement
  } else if (Component) {
    element = <Component {...props} />
  } else if (currentPath.get('path') === 'signup') {
    return signupRouter(currentPath, uri)
  } else {
    element = elementForPath(currentPath, uri)
  }

  return {
    componentAtTop: {
      element,
      hideBack: true,
      hideNavBar: true,
    },
    parseNextRoute: loginRouter,
  }
}

export default {
  parseRoute: loginRouter,
}

// Hack for now, until new routing is finished
const _DispatchHOC = ({dispatch, children: Child}) => (
  <Child dispatch={dispatch} />
)

const DispatchHOC = connect(
  () => ({}),
  (dispatch: any) => ({dispatch}),
)(_DispatchHOC)
