function drawRoundedRectangle(context, x, y, w, h, r) {
  context.beginPath();
  context.moveTo(x + r, y);
  context.lineTo(x + w - r, y);
  context.arc(x + w - r, y + r, r, 3 * Math.PI/ 2, 2 * Math.PI);
  context.lineTo(x + w, y + h - r);
  context.arc(x + w - r, y + h - r, r, 0, Math.PI / 2);
  context.lineTo(x + r, y + h);
  context.arc(x + r, y + h - r, r, Math.PI / 2, Math.PI);
  context.lineTo(x, y + r);
  context.arc(x + r, y + r, r, Math.PI, 3 * Math.PI / 2);
  context.stroke();
}
