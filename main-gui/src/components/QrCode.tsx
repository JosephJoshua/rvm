import QRCode from 'qrcode';
import { createEffect, onMount } from 'solid-js';

export type QrCodeProps = {
  text: string;
};

const QrCode = (props: QrCodeProps) => {
  let canvas: HTMLCanvasElement | undefined;

  createEffect(() => {
    if (canvas === undefined) return;
    QRCode.toCanvas(canvas, props.text, {
      errorCorrectionLevel: 'H',
    });
  });

  onMount(() => {
    if (canvas === undefined) return;
    QRCode.toCanvas(canvas, props.text, {
      errorCorrectionLevel: 'H',
    });
  });

  return <canvas ref={canvas} />;
};

export default QrCode;
