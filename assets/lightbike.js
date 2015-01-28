var conf = {
  gridWidth: 100,
  bikeHeight: 3,
  bikeWidth: 3
}

function paintLightbikePoint(ctx, c, x, y) {
  ctx.beginPath()
  ctx.moveTo(x, y)
  ctx.lineTo(x + c.bikeWidth, y)
  ctx.lineTo(x + c.bikeWidth, y + c.bikeHeight)
  ctx.lineTo(x, y + c.bikeHeight)
  ctx.closePath()
  ctx.fill()
}

function paintLightbike(ctx, c, bike) {
  ctx.fillStyle = bike.color
  bike.paths.forEach(function(point) {
    // top left
    var x, y
    x = (point % c.gridWidth) * c.bikeWidth
    y = Math.floor(point / c.gridWidth) * c.bikeHeight
    paintLightbikePoint(ctx, c, x, y)
  })
}

function paint(ctx, c, bike) {
  ctx.clearRect(0, 0, 200, 200)
  paintLightbike(ctx, c, bike)
}

var ctx = document.getElementById('canvas').getContext('2d')
function url(s) {
    var l = window.location
    return ((l.protocol === "https:") ? "wss://" : "ws://") + l.hostname + (((l.port != 80) && (l.port != 443)) ? ":" + l.port : "") + s
}
var ws = new WebSocket(url('/lightbike.ws'))
ws.onopen = function() {
  ws.onmessage = function(e) {
    try {
      var bike = JSON.parse(e.data)
      paint(ctx, conf, bike)
    } catch (err) {
      console.log(err)
    }
  }
  var state = {left: false, right: false, up: false, down: false}
  var mappings = [
    {
      button: 'left',
      code: 37
    },
    {
      button: 'right',
      code: 39
    },
    {
      button: 'up',
      code: 38
    },
    {
      button: 'down',
      code: 40
    }
  ]
  mappings.forEach(function(m) {
    document.addEventListener('keyup', function(e) {
      if (e.keyCode != m.code) {
        return
      }
      if (state[m.button] == false) {
	return
      }
      state[m.button] = false
      ws.send(JSON.stringify({button: m.button, pressed: false}))
    })
    document.addEventListener('keydown', function(e) {
      if (e.keyCode != m.code) {
        return
      }
      if (state[m.button] == true) {
	return
      }
      state[m.button] = true
      ws.send(JSON.stringify({button: m.button, pressed: true}))
    })
  })
}
