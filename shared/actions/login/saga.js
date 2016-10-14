// @flow
import * as Constants from '../../constants/types/more'
import {call, put, take, race, select} from 'redux-saga/effects'
import {isMobile} from '../../constants/platform'
import {loginLoginRpcChannelMap} from '../../constants/types/flow-types'
import {createChannelMap, putOnChannelMap, singleFixedChannelConfig, getChannel, closeChannelMap, takeFromChannelMap} from '../util/saga'

import type {SagaGenerator, ChannelConfig, ChannelMap} from '../../constants/types/saga'
import type {DeviceType} from '../../constants/types/more'

// Device Provisioning Sagas

// Here you are the existing device
function startFromComputer (typeOfNewDevice: DeviceType) {
  switch (typeOfNewDevice) {
    case 'mobile':
      return provisionComputerPhone()
  }
  return provisionComputerComputer()
}

function startFromPhone (typeOfNewDevice: DeviceType) {
  switch (typeOfNewDevice) {
    case 'mobile':
      return provisionPhonePhone()
  }
  return provisionComputerPhone()
}

function * provisionComputerComputer () {
}

function * provisionComputerPhone () {
}

function * provisionPhonePhone () {
}

// Here you are the new device
function * newComputerProvisionFromPhone () {
  // Generate QR code or enter verification code
}

function * newPhoneProvisionFromComputer () {
  // read QR code or display verification code
}

function * newPhoneProvisionFromPhone () {
  // generate QR code or read QR code or enter verification code or display verifications code
}

// Login Sagas
function * startLogin (isMobile: boolean) {
  const {payload: {usernameOrEmail}} = (yield take(Constants.startLogin)): any
  const deviceType: DeviceType = isMobile ? 'mobile' : 'desktop'

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
  ])

  const loginChanMap = loginLoginRpcChannelMap(
    param: {
      deviceType,
      usernameOrEmail: usernameOrEmail,
      clientType: CommonClientType.gui,
    }
  )

  const {response: usernameOrEmailResponse} yield takeFromChannelMap(loginChanMap, 'keybase.1.loginUi.getEmailOrUsername')
  usernameOrEmailResponse.result(usernameOrEmailResponse)

  console.log('killme: replied to usernameOrEmail')

  yield race({
    passphraseInput: getChannel(loginChanMap, 'keybase.1.secretUi.getPassphrase'),
  })

}

function * firstTimeOnDevice () {
}

// Forgot username/passphrase
function * forgotUsernamePassphraseSaga () {
}

function * recoverWithPaperKey () {
}

function * recoverWithAnotherDevice () {
}

function * resetAccountViaEmail () {
}


