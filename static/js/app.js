// /static/js/app.js
// ES module version of your SweetAlert2 helper.

export function Prompt() {
  // Toast notification
  function toast({ msg = "", icon = "success", position = "top-end" } = {}) {
    const Toast = Swal.mixin({
      toast: true,
      position,
      icon,
      showConfirmButton: false,
      timer: 3000,
      timerProgressBar: true,
      didOpen: (el) => {
        el.onmouseenter = Swal.stopTimer;
        el.onmouseleave = Swal.resumeTimer;
      },
    });
    Toast.fire({ title: msg });
  }

  // Simple success dialog
  function success({ msg = "", title = "", footer = "" } = {}) {
    Swal.fire({
      icon: "success",
      title,
      text: msg,
      footer,
    });
  }

  // Simple error dialog
  function error({ msg = "", title = "", footer = "" } = {}) {
    Swal.fire({
      icon: "error",
      title,
      text: msg,
      showConfirmButton: true,
      footer,
    });
  }

  // Fully custom dialog with optional callback
  async function custom({
    icon = "",
    msg = "",
    title = "",
    showConfirmButton = true,
    willOpen,
    callback,
  } = {}) {
    const result = await Swal.fire({
      icon,
      title,
      html: msg,
      backdrop: false,
      inputAutoFocus: false,
      showCancelButton: true,
      showConfirmButton,
      willOpen: () => {
        if (typeof willOpen === "function") willOpen();
      },
      preConfirm: () => {
        const start = document.getElementById("start")?.value ?? "";
        const end = document.getElementById("end")?.value ?? "";
        return { start, end };
      },
    });

    if (typeof callback === "function") {
      if (result.isConfirmed) callback(result.value);
      else callback(false);
    }

    return result; // allow awaiting/inspection by callers
  }

  return { toast, success, error, custom };
}

// Ready-to-use instance you can import directly
export const prompty = Prompt();
