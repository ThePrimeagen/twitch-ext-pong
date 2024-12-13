/** @type {HTMLCanvasElement} */
const canvas = document.getElementById("gameCanvas");
const ctx = canvas.getContext("2d");

// Game dimensions
const GAME_WIDTH = 800;
const GAME_HEIGHT = 600;
const BORDER_WIDTH = 10;

// Draw game border
function drawGameBorder() {
    ctx.strokeStyle = '#fff';
    ctx.lineWidth = BORDER_WIDTH;
    ctx.strokeRect(
        BORDER_WIDTH / 2,
        BORDER_WIDTH / 2,
        GAME_WIDTH - BORDER_WIDTH,
        GAME_HEIGHT - BORDER_WIDTH
    );
}

// Initialize game
drawGameBorder();
