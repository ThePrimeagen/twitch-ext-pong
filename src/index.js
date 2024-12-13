/** @type {HTMLCanvasElement} */
const canvas = document.getElementById("gameCanvas");
const ctx = canvas.getContext("2d");

// Game dimensions and settings
const ASPECT_RATIO = 16/9;
const BORDER_WIDTH = 10;
const MIN_PADDING = 20;

// Paddle settings
const PADDLE_WIDTH = 20;
const PADDLE_HEIGHT = 100;
const PADDLE_COLOR = '#fff';

// Draw paddle at specific coordinates
function drawPaddle(x, y) {
    ctx.fillStyle = PADDLE_COLOR;
    ctx.fillRect(x, y, PADDLE_WIDTH, PADDLE_HEIGHT);
}

// Resize canvas to fit window
function resizeCanvas() {
    const maxWidth = window.innerWidth - MIN_PADDING * 2;
    const maxHeight = window.innerHeight - MIN_PADDING * 2;

    // Calculate size maintaining aspect ratio
    let width = maxWidth;
    let height = width / ASPECT_RATIO;

    if (height > maxHeight) {
        height = maxHeight;
        width = height * ASPECT_RATIO;
    }

    canvas.width = width;
    canvas.height = height;

    // Draw game elements
    drawGameBorder();
    drawPaddles();
}

// Draw game border
function drawGameBorder() {
    ctx.strokeStyle = '#fff';
    ctx.lineWidth = BORDER_WIDTH;
    ctx.strokeRect(
        BORDER_WIDTH / 2,
        BORDER_WIDTH / 2,
        canvas.width - BORDER_WIDTH,
        canvas.height - BORDER_WIDTH
    );
}

// Draw both paddles at initial positions
function drawPaddles() {
    // Left paddle
    drawPaddle(
        BORDER_WIDTH * 2,
        (canvas.height - PADDLE_HEIGHT) / 2
    );

    // Right paddle
    drawPaddle(
        canvas.width - BORDER_WIDTH * 2 - PADDLE_WIDTH,
        (canvas.height - PADDLE_HEIGHT) / 2
    );
}

// Initialize and handle resize
window.addEventListener('resize', resizeCanvas);
resizeCanvas();
