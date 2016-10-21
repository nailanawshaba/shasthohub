// @flow

import * as Constants from '../../constants/login'
import HiddenString from '../../util/hidden-string'

function startLogin (usernameOrEmail: string): Constants.StartLogin {
  return {type: Constants.startLogin, payload: {usernameOrEmail}}
}

function relogin (usernameOrEmail: string, passphrase: HiddenString): Constants.Relogin {
  if (typeof passphrase === 'string') {
    console.error('passphrase is a string, should be a hidden string!')
    throw new Error('DEV: passphrase is a string, should be a hidden string!')
  }
  return {type: Constants.relogin, payload: {usernameOrEmail, passphrase}}
}

function submitPassphrase (passphrase: HiddenString, storeSecret: boolean): Constants.SubmitPassphrase {
  if (typeof passphrase === 'string') {
    console.error('passphrase is a string, should be a hidden string!')
    throw new Error('DEV: passphrase is a string, should be a hidden string!')
  }
  return {type: Constants.submitPassphrase, payload: {passphrase, storeSecret}}
}

function onBack (): Constants.OnBack {
  return {type: Constants.onBack, payload: {}}
}

function onWont (): Constants.OnWont {
  return {type: Constants.onWont, payload: {}}
}

function someoneElse (): Constants.SomeoneElse {
  return {type: Constants.someoneElse, payload: {}}
}

function selectDeviceId (deviceId: string): Constants.SelectDeviceId {
  return {type: Constants.selectDeviceId, payload: {deviceId}}
}

function chooseGPGMethod (exportKey: boolean): Constants.ChooseGPGMethod {
  return {type: Constants.chooseGPGMethod, payload: {exportKey}}
}

function submitDeviceName (deviceName: string): Constants.SubmitDeviceName {
  return {type: Constants.submitDeviceName, payload: {deviceName}}
}

function onFinish (): Constants.OnFinish {
  return {type: Constants.onFinish, payload: {}}
}

function qrScanned (phrase: string): Constants.QrScanned {
  return {type: Constants.qrScanned, payload: {phrase}}
}

function provisionTextCodeEntered (phrase: string): Constants.ProvisionTextCodeEntered {
  return {type: Constants.provisionTextCodeEntered, payload: {phrase}}
}

function setRevokedSelf (revoked: string) {
  return {type: Constants.setRevokedSelf, payload: revoked}
}

function setDeletedSelf (deletedUsername: string) {
  return {type: Constants.setDeletedSelf, payload: deletedUsername}
}

function setLoginFromRevokedDevice (error: string) {
  return {type: Constants.setLoginFromRevokedDevice, payload: error}
}

function setCodePageMode (mode: Constants.Mode) {
  return {type: Constants.setCodeMode, payload: mode}
}

export {
  chooseGPGMethod,
  onBack,
  onFinish,
  onWont,
  provisionTextCodeEntered,
  qrScanned,
  relogin,
  selectDeviceId,
  setCodePageMode,
  setDeletedSelf,
  setLoginFromRevokedDevice,
  setRevokedSelf,
  someoneElse,
  startLogin,
  submitDeviceName,
  submitPassphrase,
}
