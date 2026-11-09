const storage_key = "NAV_STATE";

const getStorageKey = (id) => `${storage_key}_${id}`;

function initializeNavState() {
  const nodes = document.querySelector(".nav__list").querySelectorAll("details.schema");;

  var persistantState = {};

  for (var i = 0; i < nodes.length; i++) {
    const id = nodes[i].id;
    const state = i === 0

    if (state) {
      persistantState[id] = state;
      localStorage.setItem(getStorageKey(id), 1);
    } else {
      localStorage.removeItem(getStorageKey(id));
    }
  }

  ensureNavOpenState(persistantState);
}

function updateNavState() {
  const nodes = document.querySelector(".nav__list").querySelectorAll("details.schema");;

  var persistantState = {};

  for (var i = 0; i < nodes.length; i++) {
    const id = nodes[i].id;
    const state = localStorage.getItem(getStorageKey(id));

    if (!!state) {
      persistantState[id] = state;
    }
  }

  ensureNavOpenState(persistantState);
}

function trackNavStateChange(el) {
  if (!el.open) {
    localStorage.setItem(getStorageKey(el.id), 1);
  } else {
    localStorage.removeItem(getStorageKey(el.id));
  }
}

function ensureNavOpenState(persistantState) {
  for (var state in persistantState) {
    document.getElementById(state).open = persistantState[state];
  }
}

document.addEventListener('DOMContentLoaded', () => {
  document.body.addEventListener('htmx:afterSwap', (evt) => {
    if (evt.target.id === "nav") {
      updateNavState();
    }
  });

  initializeNavState();
});
