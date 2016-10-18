// @flow

import React, {Component} from 'react'
import _ from 'lodash'
import {Box, Avatar} from '../common-adapters'
import {InfiniteLoader, Grid, List, WindowScroller, CellMeasurer, defaultCellMeasurerCellSizeCache} from 'react-virtualized'
import {globalStyles, globalColors} from '../styles'
import ReactList from 'react-list'

function ChatMessage ({username, messageLengths, i, style}): {username: string} {
  if (messageLengths == null) {
    console.log('null message')
    return null
  }
  return (
    <Box style={{...globalStyles.flexBoxRow, ...style}}>
      <Avatar size={24} username={username}/>
      <Box style={globalStyles.flexBoxColumn}>
      <Box>{username} - {i}</Box>
      {messageLengths.map((l, i) => (
        <Box key={i} style={{backgroundColor: globalColors.grey, height: 16, marginBottom: 8, width: l * 10}} />
      ))}
      </Box>
    </Box>
  )
}

function LoadingMessage ({style}): {} {
  return (
    <Box style={{...globalStyles.flexBoxRow, ...style}}>
      <Box>--- Loading More Messages ---</Box>
    </Box>
  )
}

const usernames = ['marcopolo', 'cecileb']

class TimeBasedCellSizeCache extends defaultCellMeasurerCellSizeCache{
  updateLoadedMessages (newLoadedMessages) {
    this.loadedMessages = newLoadedMessages
  }

  getRowHeight (index) {
    let id
    // loader message
    if (index === 0) {
      id = 0
    } else {
      // minus one because loader message is there
      const m = this.loadedMessages[index - 1]
      id = m && m.timestamp
      if (id == null) {
        console.warn('id is null for index:', index - 1)
      }
    }
    return super.getRowHeight(id)
  }

  setRowHeight (index, height) {
    let id
    // loader message
    if (index === 0) {
      console.warn('setting height of cell 0', height)
      id = 0
    } else {
      // minus one because loader message is there
      const m = this.loadedMessages[index - 1]
      id = m && m.timestamp
      if (id == null) {
        console.warn('id is null for index:', index - 1)
      }
    }
    super.setRowHeight(id, height)
  }
}

class ChatDemo extends Component {
  constructor () {
    super()
    this.state = {
      loadedMessages: [],
      scrollTop: 0,
    }

    this._cellCache = new TimeBasedCellSizeCache({uniformColumnWidth: true})
    this._cellCache.updateLoadedMessages([])
    window._cellCache = this._cellCache
    this._stoppedMoving = _.debounce(() => {
      this.state.onStoppedMoving && this.state.onStoppedMoving()
      this.setState({moving: false, onStoppedMoving: null})
    }, 1e3)
    this._throttledOnScroll = _.throttle(this._handleOnScroll.bind(this), 100)
    this._cm = null
  }

  _genRandomMessage (i: number) {
    const username = usernames[_.random(0, usernames.length - 1)]
    const messageLengths = _.range(0, _.random(1, 5)).map(i => _.random(5, 18))
    return {username, messageLengths, i, timestamp: Date.now() + i}
  }

  componentWillMount () {
    if (!this.state.loadedMessages.length) {
      this._getMoreMessages(0, 20, 30).then(() => {
        window._cm = this._cellMeasurer
      })

    }
  }

  _prependMoreMessages (count, timeout = 3e3) {
    let loadedMessages = [...this.state.loadedMessages]
    loadedMessages.unshift.apply(loadedMessages, _.range(0, count).map(() => undefined))
    this._getMoreMessages(0, count, timeout, loadedMessages)
  }

  _getMoreMessages (startIndex, endIndex, timeout = 3e3, loadedMessages = null) {
    loadedMessages = [...(loadedMessages || this.state.loadedMessages)]
    if (this.state.fetchingMessages) {
      console.log('bailing on getting more messages', startIndex, endIndex)
      return
    }

    this.setState({fetchingMessages: true})
    return new Promise((resolve, reject) => {
      console.log('getting more messages for', startIndex, endIndex)
      setTimeout(() => {
        console.log('done getting messages for', startIndex, endIndex)
        _.range(startIndex, endIndex).forEach(i => {
          if (!loadedMessages[i]) {
            loadedMessages[i] = this._genRandomMessage(i)
          }
        })

        const onStoppedMoving = () => {
          //this._cellMeasurer && this._cellMeasurer.resetMeasurements()
          this._cellCache.updateLoadedMessages(loadedMessages)
          this.setState({loadedMessages, fetchingMessages: false})
        }

        if (this.state.moving) {
          this.setState({onStoppedMoving})
        } else {
          onStoppedMoving()
        }
        resolve()
      }, timeout)
    })
  }

  componentDidUpdate (prevProps, prevState) {
    if (prevState.fetchingMessages && !this.state.fetchingMessages) {
      // 1 - 11 because you are prepending 10, and there is a loading message.
      const scrollTop = this.state.scrollTop + _.range(1,11).map(index => this._cellMeasurer.getRowHeight({index})).reduce((acc, h) => acc + h, 0)
      this.setState({scrollTop})
    }
  }

  _renderWithReactList () {
    return (
      <ReactList
        useTranslate3d={true}
        itemRenderer={(i, key) => <ChatMessage key={key} {...this.state.loadedMessages[i]} />}
        length={this.state.loadedMessages.length}
        type='uniform' />
    )
  }

  _handleOnScroll ({clientHeight, scrollHeight, scrollTop}) {
    this.setState({scrollTop, moving: true})
    this._stoppedMoving()
    if (scrollTop === 0) {
      this._prependMoreMessages(10)
    }
  }

  _indexToKey (i) {
    return this.state.loadedMessages[i].timestamp
  }

  _renderWithCellMeasurer () {
    console.log('moving?', this.state.moving)
    const rowRenderer = ({index: i, style, key}) => this._isRowLoaded({index: i}) ? <ChatMessage style={style} key={this._indexToKey(i)} {...this.state.loadedMessages[i]} /> : <Box />
    const loadingRenderer = ({index, style, key, ...rest}) => index === 0 ? (<LoadingMessage style={style} key={key || index} {...rest} />) : rowRenderer({index: index - 1, style, key, ...rest})
    const rowCount = this.state.loadedMessages.length
    const countWithLoading = rowCount + 1
    return (
      <CellMeasurer
        cellRenderer={({rowIndex, ...rest}) => loadingRenderer({index: rowIndex, ...rest})}
        columnCount={1}
        cellSizeCache={this._cellCache}
        ref={r => this._cellMeasurer = r}
        rowCount={countWithLoading}
      >
        {({getRowHeight, }) => (
          <List
            height={500}
            width={400}
            scrollTop={this.state.scrollTop}
            onScroll={this._throttledOnScroll}
            rowCount={countWithLoading}
            rowHeight={getRowHeight}
            rowRenderer={loadingRenderer}
           />
        )}
        </CellMeasurer>
    )
  }

  _isRowLoaded ({index}) {
    return this.state.loadedMessages[index] != null
  }


  // Further optimization is to make the CellMeasurer's cache be smarter about prepending things
  render () {
    return (
      <Box style={{height: 500}}>
        {false && this.state.loadedMessages.map((m, i) => <ChatMessage key={i} {...m} />)}
        {this._renderWithCellMeasurer()}
      </Box>
    )
  }

}

const chatMessageMap = {
  component: ChatMessage,
  mocks: {
    'Normal': {username: 'marcopolo', messageLengths: [8, 3, 5, 9, 9, 3, 9]},
  },
}

const chatDemo = {
  component: ChatDemo,
  mocks: {
    'Normal': {},
  },
}

export default {
  ChatMessage: chatMessageMap,
  ChatDemo: chatDemo,
}
