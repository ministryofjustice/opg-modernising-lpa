export default function hasReturnToTaskListAndSaveButtons() {
    const saveOrReturn = document.querySelector('[data-module="app-save-or-return"]')

    if (saveOrReturn?.children.length !== 2) {
        return false
    }

    for (let element of saveOrReturn.children) {
        if (!['A', 'BUTTON'].includes(element.tagName)) {
            return false
        }
    }

    return true
}
