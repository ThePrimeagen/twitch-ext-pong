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

// Constants for paddle movement
const PADDLE_SPEED = 10;
const MIN_PADDLE_Y = BORDER_WIDTH;
let maxPaddleY = canvas.height - BORDER_WIDTH - PADDLE_HEIGHT;

// WebSocket connection
const ws = new WebSocket('ws://localhost:42069/ws');
let gameState = {
    leftPaddle: 0,
    rightPaddle: 0
};

// WebSocket event handlers
ws.onopen = () => {
    console.log('🦍 CONNECTED TO STRONK SERVER 🦍');
};

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);

    switch (message.type) {
        case 'initial_state':
            gameState = message.payload;
            drawGame();
            break;
        case 'paddle_update':
            const pos = message.payload;
            if (pos.side === 'left') {
                gameState.leftPaddle = pos.y;
            } else {
                gameState.rightPaddle = pos.y;
            }
            drawGame();
            break;
    }
};

ws.onerror = (error) => {
    console.error('🦍 WEBSOCKET ERROR:', error);
};

ws.onclose = () => {
    console.log('🦍 DISCONNECTED FROM SERVER 🦍');
};

// Draw paddle at specific coordinates
function drawPaddle(x, y) {
    ctx.fillStyle = PADDLE_COLOR;
    ctx.fillRect(x, y, PADDLE_WIDTH, PADDLE_HEIGHT);
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

// Draw both paddles based on game state
function drawPaddles() {
    // Left paddle
    drawPaddle(
        BORDER_WIDTH * 2,
        gameState.leftPaddle
    );

    // Right paddle
    drawPaddle(
        canvas.width - BORDER_WIDTH * 2 - PADDLE_WIDTH,
        gameState.rightPaddle
    );
}

// Draw complete game state
function drawGame() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    drawGameBorder();
    drawPaddles();
}

// Handle paddle movement
function movePaddle(direction) {
    const currentY = gameState.leftPaddle;  // All players control left paddle
    let newY = currentY;

    if (direction === 'up') {
        newY = Math.max(MIN_PADDLE_Y, currentY - PADDLE_SPEED);
    } else if (direction === 'down') {
        newY = Math.min(maxPaddleY, currentY + PADDLE_SPEED);
    }

    if (newY !== currentY) {
        // Send paddle position to server
        const updateMsg = {
            type: 'paddle_update',
            payload: {
                side: 'left',  // All players on left team
                y: newY
            }
        };
        ws.send(JSON.stringify(updateMsg));
    }
}

// Handle keyboard events
document.addEventListener('keydown', (event) => {
    switch (event.key) {
        case 'ArrowUp':
            movePaddle('up');
            break;
        case 'ArrowDown':
            movePaddle('down');
            break;
    }
});

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

    // Update max paddle Y position
    maxPaddleY = canvas.height - BORDER_WIDTH - PADDLE_HEIGHT;

    // Draw game elements
    drawGame();
}

// Initialize and handle resize
window.addEventListener('resize', resizeCanvas);
resizeCanvas();
