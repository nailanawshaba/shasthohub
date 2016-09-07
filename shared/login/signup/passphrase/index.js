// @flow
import React, {Component} from 'react'
import Render from './index.render'
import {bindActionCreators} from 'redux'
import {checkPassphrase, restartSignup} from '../../../actions/signup'
import {connect} from 'react-redux'
import HiddenString from '../../../util/hidden-string'

import type {Props as RenderProps} from './index.render'
import type {TypedState} from '../../../constants/reducer'

type State = {
  pass1: string,
  pass2: string
}

type Props = {
  passphraseError: ?HiddenString,
  checkPassphrase: (pass1: string, pass2: string) => void,
  restartSignup: () => void,
}

class PassphraseForm extends Component<void, Props, State> {
  state: State;

  constructor () {
    super()

    this.state = {
      pass1: '',
      pass2: '',
    }
  }

  render () {
    return (
      <Render
        passphraseError={this.props.passphraseError}
        pass1={this.state.pass1}
        pass1Update={pass1 => this.setState({pass1})}
        pass2={this.state.pass2}
        pass2Update={pass2 => this.setState({pass2})}
        onSubmit={() => this.props.checkPassphrase(this.state.pass1, this.state.pass2)}
        onBack={this.props.restartSignup}
        />
    )
  }
}

const F: Class<Component<void, {}, State>> = connect(
  (state: TypedState, ownProps: {}): {passphraseError: ?HiddenString} => {
    return {
      passphraseError: null,
      checkPassphrase: (pass1, pass2) => {},
    }
  },
  (dispatch: () => void, ownProps: {}): {} => ({
    restartSignup: () => {},
  })
)(PassphraseForm)

export default F
