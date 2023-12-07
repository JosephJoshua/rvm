import { Show, createSignal, onCleanup, onMount } from 'solid-js';
import { BarcodeDetector } from 'barcode-detector';
import axios, { AxiosError } from 'axios';
import LoadingIndicator from '../components/LoadingIndicator';
import Modal from '../components/Modal';
import { auth } from '../lib/firebase';

const QR_CODE_PREFIX = 'greenwaste-rvm/transaction/';

const Scan = () => {
  const barcodeDetector = new BarcodeDetector({ formats: ['qr_code'] });

  let video: HTMLVideoElement | undefined;
  let detectionInterval: number | undefined;

  const [isLoading, setIsLoading] = createSignal(false);
  const [isInvalidCode, setIsInvalidCode] = createSignal(false);

  const [transactionPoints, setTransactionPoints] = createSignal<
    number | undefined
  >(undefined);

  const handleEndTransaction = async (transactionId: string) => {
    const uid = auth.currentUser?.uid;
    if (uid === undefined) return;

    setIsLoading(true);

    try {
      const url = new URL(
        `/transactions/${transactionId}/end`,
        import.meta.env.VITE_BACKEND_URL,
      );

      const response = await axios.postForm(
        url.href,
        {
          user_id: uid,
        },
        {
          headers: {
            Authorization: `Bearer ${import.meta.env.VITE_BACKEND_TOKEN}`,
          },
        },
      );

      const points = response.data;

      setTransactionPoints(points);
      clearInterval(detectionInterval);
    } catch (e) {
      if (e instanceof AxiosError) {
        setIsInvalidCode(true);
      }

      throw e;
    } finally {
      setIsLoading(false);
    }
  };

  const handleDetect = () => {
    if (video?.srcObject == null) return;
    if (isLoading() || isInvalidCode()) return;

    barcodeDetector.detect(video).then((codes) => {
      for (const code of codes) {
        if (code.rawValue.startsWith(QR_CODE_PREFIX)) {
          const transactionId = code.rawValue.split(QR_CODE_PREFIX)[1];
          if (transactionId.length === 0) continue;

          handleEndTransaction(transactionId);
        }
      }
    });
  };

  const handleTryAgain = () => {
    setIsInvalidCode(false);
  };

  onMount(() => {
    navigator.mediaDevices
      .getUserMedia({
        video: true,
        audio: false,
        preferCurrentTab: true,
      })
      .then((stream) => {
        if (video === undefined) return;
        video.srcObject = stream;
      });

    detectionInterval = setInterval(handleDetect, 300);
  });

  onCleanup(() => {
    clearInterval(detectionInterval);
  });

  return (
    <>
      <div class="min-h-screen flex flex-col justify-center items-center p-8 gap-6">
        <div class="w-full max-w-[640px] relative">
          <video autoplay ref={video} class="rounded-md w-full" />

          <Show when={isLoading()}>
            <div class="bg-black/70 absolute inset-0 z-20 flex justify-center items-center rounded-md">
              <LoadingIndicator class="w-16 h-16" />
            </div>
          </Show>
        </div>

        <a
          href="/"
          class="transition duration-300 bg-red-600 text-white px-12 py-2 rounded-md font-medium text-lg hover:-translate-y-0.5"
        >
          Cancel
        </a>
      </div>

      <Modal show={isInvalidCode()}>
        <div class="p-8">
          <h2 class="text-2xl font-semibold mb-4">Invalid QR Code</h2>

          <p class="mb-6">
            The QR code you provided is invalid. Please try again.
          </p>

          <div class="flex justify-center items-center">
            <button
              type="button"
              class="transition duration-300 bg-red-600 text-white px-12 py-2 rounded-md font-medium text-lg hover:-translate-y-0.5"
              onClick={handleTryAgain}
            >
              OK
            </button>
          </div>
        </div>
      </Modal>

      <Modal show={transactionPoints() !== undefined} showCloseButton={false}>
        <div class="p-8">
          <h2 class="text-2xl font-semibold mb-4">Transaction completed</h2>

          <p class="mb-6">
            You have successfully completed your transaction.{' '}
            <span class="text-green-600">{transactionPoints()} points</span> has
            been added to your account. Thank you for using GreenWaste RVM.
          </p>

          <div class="flex justify-center items-center">
            <a
              href="/"
              class="transition duration-300 bg-green-600 text-white px-12 py-2 rounded-md font-medium text-lg hover:-translate-y-0.5"
            >
              Done
            </a>
          </div>
        </div>
      </Modal>
    </>
  );
};

export default Scan;
