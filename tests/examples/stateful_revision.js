var counter = getState('counter')
counter = counter ? counter : 0

setState('counter', counter + 1)

var result = getState('counter')