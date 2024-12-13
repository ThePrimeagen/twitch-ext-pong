/** @type {HTMLCanvasElement} */
const canvas = document.getElementById("gameCanvas");
const ctx = canvas.getContext("2d");

// Game dimensions and settings
const ASPECT_RATIO = 4/3;
const BORDER_WIDTH = 10;
const MIN_PADDING = 50;

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

    // Redraw game border
    drawGameBorder();
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

// Initialize and handle resize
window.addEventListener('resize', resizeCanvas);
resizeCanvas();
