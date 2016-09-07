// @flow
import React, {Component} from 'react'
import Render from './index.render'
import {connect} from 'react-redux'
import {sawPaperKey} from '../../../actions/signup'

import type {TypedState} from '../../../constants/reducer'

class Success extends Component {
  render () {
    return (
      <Render
        title={this.props.title}
        paperkey={this.props.paperkey}
        waiting={this.props.waiting}
        onFinish={this.props.onFinish}
        onBack={this.props.onBack}
        />
    )
  }
}

export default connect(
  (state: TypedState) => ({
    paperkey: state.signup.paperkey,
    waiting: state.signup.waiting,
  }),
  dispatch => ({
    onFinish: () => dispatch(sawPaperKey()),
    onBack: () => {},
  })
)(Success)
