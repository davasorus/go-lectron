import {QRService} from "../bindings/changeme";

const qrTextElement = document.getElementById('text')! as HTMLInputElement;
const qrImageElement = document.getElementById('qrcode')! as HTMLImageElement;
const generateButton = document.getElementById('generateButton')! as HTMLButtonElement;

generateButton.addEventListener('click', async () => {
    const text = qrTextElement.value;
    if (!text) {
        alert('Please enter some text');
        return;
    }
    try {
        qrImageElement.src = await QRService.URL(text, 256);
    } catch (err) {
        console.error('Failed to generate QR code:', err);
    }
});