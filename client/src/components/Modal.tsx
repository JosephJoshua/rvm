import { ParentProps, Show } from 'solid-js';

export type ModalProps = ParentProps & {
  show?: boolean;
  onClose?: () => void;
  showCloseButton?: boolean;
};

const Modal = (props: ModalProps) => {
  return (
    <Show when={props.show}>
      <div class="z-40 bg-black/70 fixed inset-0" />

      <div
        tabindex="-1"
        class="overflow-y-auto overflow-x-hidden fixed top-0 right-0 left-0 z-50 flex justify-center items-center w-full md:inset-0 h-[calc(100%-1rem)] max-h-full"
      >
        <div class="relative p-4 w-full max-w-md max-h-full">
          <div class="relative bg-white rounded-lg shadow-md mx-auto">
            <Show when={props.showCloseButton ?? true}>
              <button
                type="button"
                class="absolute top-3 end-2.5 text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm w-8 h-8 ms-auto inline-flex justify-center items-center"
                onClick={() => props.onClose?.()}
              >
                <svg
                  class="w-3 h-3"
                  aria-hidden="true"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 14 14"
                >
                  <path
                    stroke="currentColor"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="m1 1 6 6m0 0 6 6M7 7l6-6M7 7l-6 6"
                  />
                </svg>

                <span class="sr-only">Close modal</span>
              </button>
            </Show>

            {props.children}
          </div>
        </div>
      </div>
    </Show>
  );
};

export default Modal;
