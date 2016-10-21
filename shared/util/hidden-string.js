// @flow

const hiddenStringName = '[HiddenString]'
// HiddenString tries to wrap a string value to prevent it from being easily
// output as a string to log, file or console
class HiddenString {
  _value: () => string;

  constructor (stringValue: string) {
    this._value = () => stringValue
  }

  toString (): string {
    return hiddenStringName
  }

  stringValue (): string {
    return this._value()
  }
}

function throwIfNotHidden (x: any): HiddenString {
  if (x.toString() !== hiddenStringName) {
    throw new Error('This should be in a hidden string')
  }
  return x
}

export {
  throwIfNotHidden,
}

export default HiddenString
