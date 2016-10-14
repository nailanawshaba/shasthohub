// @flow

import * as Constants from '../../constants/login'

function startLogin (usernameOrEmail: string) {
  return {type: Constants.startLogin, payload: {usernameOrEmail}}
}

export {
  startLogin,
}
