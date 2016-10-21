// @flow
import React from 'react'
import * as Constants from '../../constants/login'
import * as ConfigConstants from '../../constants/config'
import * as routerActions from '../router'
import * as Creators from './creators'
import _ from 'lodash'
import {call, put, take, race, select} from 'redux-saga/effects'
import {takeLatest, takeEvery} from 'redux-saga'
import {isMobile} from '../../constants/platform'
import {devicesTab, loginTab} from '../../constants/tabs'
import {loginLoginRpcChannelMap} from '../../constants/types/flow-types'
import {createChannelMap, putOnChannelMap, singleFixedChannelConfig, getChannel, closeChannelMap, takeFromChannelMap} from '../../util/saga'
import {overrideLoggedInTab} from '../../local-debug'
import HiddenString from '../../util/hidden-string'
import {defaultModeForDeviceRoles, qrGenerate} from './provision-helpers'

import {loginRecoverAccountFromEmailAddressRpc, loginLoginRpc, loginLogoutRpc,
  deviceDeviceAddRpc, loginGetConfiguredAccountsRpcPromise, CommonClientType,
  ConstantsStatusCode, ProvisionUiGPGMethod, ProvisionUiDeviceType,
  PassphraseCommonPassphraseType,
} from '../../constants/types/flow-types'

import type {SagaGenerator, ChannelConfig, ChannelMap, AfterSelect} from '../../constants/types/saga'
import type {DeviceType} from '../../constants/types/more'
import type {TypedState} from '../../constants/reducer'

function _waitingForResponse (waiting: boolean) {
  return {
    type: Constants.waitingForResponse,
    payload: waiting,
  }
}

const codePageSelector = ({login: {codePage}}: TypedState) => codePage

function * passphraseFlow ({params: {type, prompt, username, retryLabel}, response}) {
  switch (type) {
    case PassphraseCommonPassphraseType.paperKey:
      yield put(routerActions.routeAppend({
        path: 'paperkey',
        props: {
          retryLabel,
        },
      }))
      break
    case PassphraseCommonPassphraseType.passPhrase:
      yield put(routerActions.routeAppend({
        path: 'passphrase',
        props: {
          prompt,
          username,
          retryLabel,
        },
      }))
      break
    default:
      yield call(
        [response, response.error],
        {
          code: ConstantsStatusCode.scnotfound,
          desc: 'Unknown getPassphrase type',
        }
      )
      return false
  }

  const {payload: {passphrase}} = ((yield take(Constants.submitPassphrase)): any)
  yield call([response, response.result], {
    passphrase: passphrase.stringValue(),
    storeSecret: true,
  })
  return true
}

function loginRpc (channelConfig, usernameOrEmail) {
  const deviceType: DeviceType = isMobile ? 'mobile' : 'desktop'
  return loginLoginRpcChannelMap(
    channelConfig,
    {param: {
      deviceType,
      usernameOrEmail,
      clientType: CommonClientType.gui,
    }}
  )
}

function * deviceNameFlow ({params: {existingDevices, errorMessage}, response}) {
  yield put(routerActions.routeAppend({
    path: 'promptDeviceName',
    props: {
      existingDevices,
      errorMessage,
    },
  }))
  const {payload: {deviceName}}: Constants.SubmitDeviceName = ((yield take(Constants.submitDeviceName)): any)

  yield call([response, response.result], deviceName)
}

function * chooseGPGMethodFlow ({response}) {
  yield put(routerActions.routeAppend({
    path: 'chooseGPGMethod',
    props: {},
  }))
  const {payload: {exportKey}}: Constants.ChooseGPGMethod = ((yield take(Constants.chooseGPGMethod)): any)

  const ourResponse = exportKey ? ProvisionUiGPGMethod.gpgImport : ProvisionUiGPGMethod.gpgSign
  yield call([response, response.result], ourResponse)
}

function * chooseDeviceFlow ({params: {devices}, response}) {
  yield put(routerActions.routeAppend({
    path: 'chooseDevice',
    props: {
      devices,
    },
  }))

  const {payload: {deviceId}}: Constants.SelectDeviceId = ((yield take(Constants.selectDeviceId)): any)
  yield call([response, response.result], deviceId)
}

function * provisionerSuccess ({response}) {
  yield call([response, response.result])
  yield call(navBasedOnLoginState)
}

function * provisioneeSuccess ({response}) {
  yield call([response, response.result])
}

function * displaySecretExchanged ({response}) {
  yield call([response, response.result])
}

function * selectGPGKey ({response}) {
  yield call([response, response.error], {
    code: ConstantsStatusCode.sckeynotfound,
    desc: 'Not supported in GUI',
  })
}

function * displayPaperKeyFlow ({params: {phrase}, response}) {
  yield put(routerActions.routeAppend({
    path: 'displayPaperKey',
    props: {
      phrase: new HiddenString(phrase),
    },
  }))

  yield take(Constants.onFinish)
  yield call([response, response.result])
}

function * displayAndPromptSecretFlow ({params: {phrase, secret, otherDeviceType}, response}) {
  const otherDeviceRole: Constants.DeviceRole = otherDeviceType === ProvisionUiDeviceType.desktop
    ? Constants.codePageDeviceRoleExistingComputer : Constants.codePageDeviceRoleExistingPhone
  yield put({type: Constants.setOtherDeviceCodeState, payload: otherDeviceRole})

  const codePage: AfterSelect<typeof codePageSelector> = ((yield select(codePageSelector)): any)
  if (codePage.myDeviceRole == null) {
    console.warn("my device role is null, can't setCodePageOtherDeviceRole. Bailing")
    return
  }

  // $FlowIssue
  yield put(Creators.setCodePageMode(defaultModeForDeviceRoles(codePage.myDeviceRole, otherDeviceRole, false)))

  yield put({type: Constants.setTextCode, payload: {textCode: new HiddenString(phrase)}})
  yield call(generateQRCode)

  yield put(routerActions.routeAppend('codePage'))

  const {qr, code}: {qr: ?Constants.QrScanned, code: Constants.ProvisionTextCodeEntered} = ((yield race({
    qr: take(Constants.qrScanned),
    code: take(Constants.provisionTextCodeEntered),
  })): any)

  const phraseToSend = qr ? qr.payload.phrase : code.payload.phrase
  yield call([response, response.result], phraseToSend)
  // TODO handle onBack
}

function * generateQRCode () {
  const codePage: AfterSelect<typeof codePageSelector> = ((yield select(codePageSelector)): any)
  const goodMode = codePage.mode === Constants.codePageModeShowCode

  if (goodMode && !codePage.qrCode && codePage.textCode) {
    yield put({type: Constants.setQRCode, payload: {qrCode: new HiddenString(qrGenerate(codePage.textCode.stringValue()))}})
  }
}

const methodsToDefaultSagas = {
  'keybase.1.provisionUi.chooseDevice': chooseDeviceFlow,
  'keybase.1.secretUi.getPassphrase': passphraseFlow,
  'keybase.1.provisionUi.DisplayAndPromptSecret': displayAndPromptSecretFlow,
  'keybase.1.provisionUi.PromptNewDeviceName': deviceNameFlow,
  'keybase.1.provisionUi.chooseGPGMethod': chooseGPGMethodFlow,
  'keybase.1.loginUi.displayPrimaryPaperKey': displayPaperKeyFlow,
  'keybase.1.provisionUi.ProvisioneeSuccess': provisioneeSuccess,
  'keybase.1.provisionUi.ProvisionerSuccess': provisionerSuccess,
  'keybase.1.provisionUi.DisplaySecretExchanged': displaySecretExchanged,
  'keybase.1.gpgUi.selectKey': selectGPGKey,
}

// Login Sagas
function * startLogin ({payload: {usernameOrEmail}}: Constants.StartLogin, isMobile: boolean) {
  yield put({
    type: Constants.setMyDeviceCodeState,
    payload: isMobile ? Constants.codePageDeviceRoleNewPhone : Constants.codePageDeviceRoleNewComputer,
  })

  const channelConfig = singleFixedChannelConfig([
    'keybase.1.loginUi.getEmailOrUsername',
    'keybase.1.provisionUi.chooseDevice',
    'keybase.1.secretUi.getPassphrase',
    'keybase.1.provisionUi.DisplayAndPromptSecret',
    'keybase.1.provisionUi.PromptNewDeviceName',
    'keybase.1.provisionUi.chooseGPGMethod',
    'keybase.1.loginUi.displayPrimaryPaperKey',
    'keybase.1.provisionUi.ProvisioneeSuccess',
    'keybase.1.provisionUi.ProvisionerSuccess',
    'keybase.1.provisionUi.DisplaySecretExchanged',
    'keybase.1.gpgUi.selectKey',
    'finished',
  ])

  const loginChanMap = ((yield call(loginRpc, channelConfig, usernameOrEmail)): any)

  yield _.map(methodsToDefaultSagas, (saga, methodName) => takeEvery(getChannel(loginChanMap, methodName), saga))

  // Handle this at any point in the future
  yield takeEvery(getChannel(loginChanMap, 'keybase.1.loginUi.getEmailOrUsername'), ({response: usernameOrEmailResponse}) => {
    usernameOrEmailResponse.result(usernameOrEmail)
  })

  const {error, params: status} = ((yield takeFromChannelMap(loginChanMap, 'finished')): any)
  yield finishLogin(error, status)
}

function * relogin ({payload: {usernameOrEmail, passphrase}}: Constants.Relogin) {
  // TODO handle device provisioning
  const channelConfig = singleFixedChannelConfig([
    'keybase.1.secretUi.getPassphrase',
    'finished',
  ])

  const loginChanMap = ((yield call(loginRpc, channelConfig, usernameOrEmail, passphrase.stringValue())): any)

  const passphraseInput = ((yield takeFromChannelMap(loginChanMap, 'keybase.1.secretUi.getPassphrase')): any)
  yield call([passphraseInput.response, passphraseInput.response.result], {
    passphrase: passphrase.stringValue(),
    storeSecret: true,
  })

  const {error, params: status} = ((yield takeFromChannelMap(loginChanMap, 'finished')): any)
  yield finishLogin(error, status)
}

function * finishLogin (error, status) {
  if (error) {
    console.log(error)
    yield put({type: Constants.loginDone, error: true, payload: error})
  } else {
    yield put({type: Constants.loginDone, payload: status})
    yield call(navBasedOnLoginState)
  }
}

function * getAccounts () {
  yield put(_waitingForResponse(true))

  try {
    const accounts = yield call(loginGetConfiguredAccountsRpcPromise)
    yield put({type: Constants.configuredAccounts, payload: {accounts}})
  } catch (error) {
    yield put({type: Constants.configuredAccounts, error: true, payload: error})
  } finally {
    yield put(_waitingForResponse(false))
  }
}

function * navBasedOnLoginState () {
  const selector = ({config: {status, extendedConfig}, login: {justDeletedSelf}}: TypedState) => ({status, extendedConfig, justDeletedSelf})

  const {status, extendedConfig, justDeletedSelf} = ((yield select(selector)): any)

  // No status?
  if (!status || !Object.keys(status).length || !extendedConfig || !Object.keys(extendedConfig).length ||
    !extendedConfig.defaultDeviceID || justDeletedSelf) { // Not provisioned?
    yield put(routerActions.navigateTo([], loginTab))
    yield put(routerActions.switchTab(loginTab))
  } else {
    if (status.loggedIn) { // logged in
      if (overrideLoggedInTab) {
        console.log('Loading overridden logged in tab')
        yield put(routerActions.switchTab(overrideLoggedInTab))
      } else {
        yield put(routerActions.switchTab(devicesTab))
      }
    } else if (status.registered) { // relogging in
      yield call(getAccounts)
      yield put(routerActions.navigateTo(['login'], loginTab))
      yield put(routerActions.switchTab(loginTab))
    } else { // no idea
      yield put(routerActions.navigateTo([], loginTab))
      yield put(routerActions.switchTab(loginTab))
    }
  }
}

function * someoneElse () {
  yield put(routerActions.routeAppend('usernameOrEmail'))
  // TODO (MM) handle on back / on submit

  const action = yield take(Constants.startLogin)
  yield call(startLogin, action)
}

function * mainSaga (): SagaGenerator<any, any> {
  yield [
    takeLatest(Constants.someoneElse, someoneElse),
    takeLatest(Constants.relogin, relogin),
    takeEvery(ConfigConstants.bootstrapped, navBasedOnLoginState),
  ]
}

export default mainSaga
