// @flow
/*
 * Screen to scan/show qrcode/text code. Goes into various modes with various options depending on if
 * you're a phone/computer and if you're the existing device or the new device
 */
import React, {Component} from 'react'
import Render from './index.render'
import * as actions from '../../../actions/login/creators'
import {connect} from 'react-redux'

import type {TypedState} from '../../../constants/reducer'
import type {Props} from './index.render'

type State = {
  enterText: string,
}

class CodePage extends Component<void, Props, State> {
  state: State;

  constructor (props: Props) {
    super(props)

    this.state = {
      enterText: '',
    }
  }

  render () {
    return (
      <Render
        enterText={this.state.enterText}
        onChangeText={enterText => this.setState({enterText})}
        onBack={this.props.onBack}
        mode={this.props.mode}
        textCode={this.props.textCode}
        qrCode={this.props.qrCode}
        myDeviceRole={this.props.myDeviceRole}
        otherDeviceRole={this.props.otherDeviceRole}
        cameraBrokenMode={this.props.cameraBrokenMode}
        setCodePageMode={this.props.setCodePageMode}
        qrScanned={this.props.qrScanned}
        setCameraBrokenMode={this.props.setCameraBrokenMode}
        textEntered={() => this.props.textEntered(this.state.enterText)}
        doneRegistering={this.props.doneRegistering}
      />
    )
  }
}

export default connect(
   ({login: {codePage: {
     mode, codeCountDown, textCode, qrCode,
     myDeviceRole, otherDeviceRole, cameraBrokenMode,
   }}}: TypedState) => ({
     mode,
     codeCountDown,
     textCode: textCode ? textCode.stringValue() : '',
     qrCode: qrCode ? qrCode.stringValue() : '',
     myDeviceRole,
     otherDeviceRole,
     cameraBrokenMode,

   }),
   (dispatch: any, ownProps: {}) => ({
     onBack: () => dispatch(actions.onBack()),
     setCodePageMode: mode => console.log('todo'),
     qrScanned: code => console.log('todo'),
     setCameraBrokenMode: broken => console.log('todo'),
     textEntered: text => console.log('todo'),
     doneRegistering: () => console.log('todo'),
   })
)(CodePage)
