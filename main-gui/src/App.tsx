import axios from 'axios';
import { Transaction } from './types/transaction';
import {
  Match,
  Show,
  Switch,
  createSignal,
  onCleanup,
  onMount,
} from 'solid-js';
import QrCode from './components/QrCode';

const App = () => {
  const [transaction, setTransaction] = createSignal<Transaction | null>(null);
  const [showQrCode, setShowQrCode] = createSignal<boolean>(false);
  const [isLoading, setIsLoading] = createSignal<boolean>(false);

  const transactionStarted = () => transaction() !== null;
  const transactionEmpty = () => transaction()?.itemCount === 0;

  const transactionQrText = () => {
    const id = transaction()?.id;
    if (id == null) return '';

    return `greenwaste-rvm/transaction/${id}`;
  };

  const handleAddItem = async () => {
    const transactionId = transaction()?.id;
    if (transactionId == null) return;

    const url = new URL(
      `/transactions/${transactionId}/items`,
      import.meta.env.VITE_BACKEND_URL,
    );

    const response = await axios.postForm(
      url.href,
      {
        item_id: 1,
      },
      {
        headers: {
          Authorization: `Bearer ${import.meta.env.VITE_BACKEND_TOKEN}`,
        },
      },
    );

    const newCount = response.data;
    setTransaction({
      id: transactionId,
      itemCount: newCount,
    });
  };

  const handleKeypress = (e: KeyboardEvent) => {
    if (e.ctrlKey && e.key === 'b') {
      handleAddItem();
      return;
    }
  };

  const handleStartTransaction = async () => {
    setIsLoading(true);

    try {
      const url = new URL('/transactions', import.meta.env.VITE_BACKEND_URL);
      const response = await axios.post(url.href, null, {
        headers: {
          Authorization: `Bearer ${import.meta.env.VITE_BACKEND_TOKEN}`,
        },
      });

      const transactionId = response.data;
      setTransaction({
        id: transactionId,
        itemCount: 0,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleCancelTransaction = () => {
    setShowQrCode(false);
    setTransaction(null);
  };

  const handleEndTransaction = async () => {
    const tr = transaction();
    if (tr === null) return;

    if (tr.itemCount === 0) {
      setTransaction(null);
      return;
    }

    setShowQrCode(true);
  };

  onMount(() => {
    document.addEventListener('keypress', handleKeypress);
  });

  onCleanup(() => {
    document.removeEventListener('keypress', handleKeypress);
  });

  return (
    <div class="flex flex-col items-center justify-between min-h-screen gap-8 py-16">
      <h1 class="text-green-600 text-4xl tracking-wide font-medium">
        GreenWaste
      </h1>

      <div class="flex flex-col items-center gap-6">
        <Show
          when={!showQrCode()}
          fallback={<QrCode text={transactionQrText()} />}
        >
          <div
            class="font-medium text-lg"
            classList={{
              visible: transactionStarted(),
              invisible: !transactionStarted(),
            }}
          >
            <Show
              when={!transactionEmpty()}
              fallback={'Please insert your plastic bottles into the machine.'}
            >
              <strong>x{transaction()?.itemCount}</strong> plastic bottles
              inserted.
            </Show>
          </div>
        </Show>

        <Switch>
          <Match when={!transactionStarted()}>
            <button
              type="button"
              class="transition duration-300 bg-green-600 text-white px-12 py-2 rounded-md font-medium text-lg min-w-[180px] hover:-translate-y-0.5"
              onClick={handleStartTransaction}
            >
              <Show when={!isLoading()} fallback={'Starting...'}>
                Start
              </Show>
            </button>
          </Match>

          <Match when={transactionStarted() && !showQrCode()}>
            <button
              type="button"
              class="transition duration-300 bg-green-600 text-white px-12 py-2 rounded-md font-medium text-lg hover:-translate-y-0.5"
              onClick={handleEndTransaction}
            >
              Done!
            </button>
          </Match>

          <Match when={transactionStarted() && showQrCode()}>
            <button
              type="button"
              class="transition duration-300 bg-red-600 text-white px-12 py-2 rounded-md font-medium text-lg hover:-translate-y-0.5"
              onClick={handleCancelTransaction}
            >
              Cancel
            </button>
          </Match>
        </Switch>
      </div>

      <div />
    </div>
  );
};

export default App;
