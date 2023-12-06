import axios from 'axios';
import { Transaction } from './types/transaction';
import { Show, createSignal } from 'solid-js';

const App = () => {
  const [transaction, setTransaction] = createSignal<Transaction | null>(null);
  const [isLoading, setIsLoading] = createSignal<boolean>(false);

  const transactionStarted = () => transaction() !== null;

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

  return (
    <div class="flex flex-col items-center justify-between min-h-screen gap-8 py-16">
      <h1 class="text-green-600 text-4xl tracking-wide font-medium">
        GreenWaste
      </h1>

      <div class="flex flex-col items-center gap-6">
        <div
          class="font-medium text-lg"
          classList={{
            visible: transactionStarted(),
            invisible: !transactionStarted(),
          }}
        >
          Please insert your plastic bottles into the machine.
        </div>

        <button
          type="button"
          class="transition duration-300 bg-green-600 text-white px-12 py-2 rounded-md font-medium text-lg hover:-translate-y-0.5"
          onClick={handleStartTransaction}
        >
          <Show when={!isLoading()} fallback={'Starting...'}>
            Start
          </Show>
        </button>
      </div>

      <div />
    </div>
  );
};

export default App;
