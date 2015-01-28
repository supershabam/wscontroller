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
var ws = new WebSocket('ws://localhost:8080/lightbike.ws')
ws.onopen = function() {
  ws.onmessage = function(e) {
    console.log('ws', e)
    try {
      var bike = JSON.parse(e.data)
      paint(ctx, conf, bike)
    } catch (err) {
      console.log(err)
    }
  }
}
