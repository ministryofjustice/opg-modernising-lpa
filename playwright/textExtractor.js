import {sanitisedPath} from "./e2e";
import fs from "fs";

// So it's usable in browser context but still testable
function browserTextExtractorLogic() {
    function parseChildrenNodes(element) {
        const isSummaryList = element.classList.contains('govuk-summary-list__row');
        const isListItem = element.tagName.toLowerCase() === 'li';
        let prependChar = '';

        if (isSummaryList) {
            prependChar = '|';
        } else if (isListItem) {
            prependChar = '•';
        }

        let textParts = [];

        function processNode(node) {
            if (node.nodeType === Node.ELEMENT_NODE) {
                if (node.tagName.toLowerCase() === 'a') {
                    textParts.push(`[LINK: ${node.textContent.trim()}](${node.href})`)
                } else if (node.tagName.toLowerCase() === 'span') {
                    textParts.push(`${node.textContent.trim()}`)
                } else {
                    Array.from(node.childNodes).forEach(processNode)
                }
            } else if (node.nodeType === Node.TEXT_NODE) {
                const textContent = node.textContent?.replace(/\s+/g, " ").trim()
                if (textContent) {
                    textParts.push(prependChar && !textParts.some(part => part.startsWith('•')) ? `${prependChar} ${textContent}` : `${textContent}`)
                }
            }
        }

        Array.from(element.childNodes).forEach(processNode);

        let finalText = textParts.join(' ');

        if (isSummaryList) {
            const childElements = Array.from(element.children);
            const parts = [];

            childElements.forEach(childEl => {
                let childText = '';
                Array.from(childEl.childNodes).forEach(node => {
                    if (node.nodeType === Node.ELEMENT_NODE) {
                        if (node.tagName.toLowerCase() === 'a') {
                            childText += `[LINK: ${node.textContent.trim()}](${node.href}) `;
                        } else {
                            childText += `${node.textContent.trim()} `;
                        }
                    } else if (node.nodeType === Node.TEXT_NODE) {
                        const textContent = node.textContent?.replace(/\s+/g, " ").trim();
                        if (textContent) {
                            childText += `${textContent} `;
                        }
                    }
                });
                parts.push(childText.trim());
            });

            finalText = '| ' + parts.join(' | ') + ' |';
        }

        return finalText.trim() + '\n\n';
    }

    function getElementText(element) {
        const tagName = element.tagName.toLowerCase();
        const type = element.getAttribute('type');
        const classList = element.classList;
        let text = '';

        if ( ['label', 'option', 'script', 'style'].includes(tagName) ||
            ['govuk-visually-hidden', 'govuk-hint', 'govuk-!-display-none', 'app-dialog', 'govuk-warning-text__icon'].some(
                cls => element.classList.contains(cls))
        ) {
            return '';
        }

        if ( ['p', 'li', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6'].includes(tagName) ||
            classList.contains('govuk-summary-list__row')) {
            return parseChildrenNodes(element);
        }

        switch (true) {
            case (classList.contains('govuk-button') ||
                tagName === 'button' ||
                (tagName === 'input' && type === 'button') ||
                (tagName === 'a' && element.getAttribute('role') === 'button')):
                const textContent = element.textContent.trim()
                text += `[BUTTON: ${textContent ? textContent : element.value}]\n\n`;
                break;

            case (tagName === 'input' && (type === 'radio' || type === 'checkbox')):
                const checked = element.checked ? '(checked)' : '(unchecked)';
                const label = document.querySelector(`label[for="${element.id}"]`);
                const labelText = label ? label.textContent.trim() : (element.value || '');
                text += `[${type.toUpperCase()} ${checked}: ${labelText}]\n\n`;
                break;

            case (tagName === 'select'):
                const options = Array.from(element.options)
                    .map(option => {
                        return (option.selected ? '> ' : '') + (option.textContent ? option.textContent.trim() : '<NO TEXT>');
                    })
                    .join('\n');
                text += `[SELECT: ${element.name || ''}]\n${options}\n\n`;
                break;

            case ((tagName === 'input' || tagName === 'textarea') && !['checkbox', 'select', 'radio', 'hidden', 'button'].includes(type)):
                const inputTextareaLabel = element.id ? document.querySelector(`label[for="${element.id}"]`) : null;
                const inputTextareaLabelText = inputTextareaLabel ? inputTextareaLabel.textContent.trim() : '';
                const inputTextareaHint = element.closest('.govuk-form-group')?.querySelector('.govuk-hint') ||
                    element.parentNode?.querySelector('.govuk-hint');

                text += inputTextareaLabelText ? `${inputTextareaLabelText}\n` : '';
                if (inputTextareaHint) {
                    text += `${inputTextareaHint.textContent.trim()}\n`;
                }
                const value = element.value || '<NO VALUE>';
                text += tagName === 'input' ? `[INPUT: ${value}]\n\n` : `[TEXTAREA: ${value}]\n\n`;
                break;

            case element.classList.contains('govuk-error-message'):
                text += `[ERROR: ${element.textContent?.replace(/\s+/g, " ").trim()}]\n\n`;
                break;

            case element.classList.contains('govuk-warning-text'):
                text += `[WARNING: ${element.textContent?.replace(/\s+/g, " ").trim()}]\n\n`;
                break;

            case tagName.toLowerCase() === 'a':
                const linkText = element.textContent ? element.textContent.replace(/\s+/g, " ").trim() : '<NO TEXT>';
                const href = element.href || '<NO HREF>';
                if (!element.closest('p, li, .govuk-summary-list__row')) {
                    text += `[LINK: ${linkText}](${href})\n\n`;
                }
                break;

            default:
                if (element.childNodes.length === 0 ||
                    (element.childNodes.length === 1 && element.childNodes[0].nodeType === Node.TEXT_NODE)) {
                    if (element.textContent && !element.parentNode.classList.contains('govuk-warning-text'))  {
                        text += `${element.textContent.trim()}\n\n`;
                    }
                }
                break;
        }

        for (const child of element.children) {
            text += getElementText(child);
        }

        return text;
    }

    const main = document.querySelector('main');
    if (!main) {
        return 'No main element found';
    }

    const mainContentText = getElementText(main);

    return `[PAGE TITLE: ${document.title}]\n\n` + mainContentText;
}


export async function extractTextFromMainAndSave(page) {
    const textContent = await page.evaluate(browserTextExtractorLogic);

    try {
        const dir = 'playwright/screenshots';
        if (!fs.existsSync(dir)){
            fs.mkdirSync(dir, { recursive: true });
        }
        const filename = `playwright/screenshots/${Date.now()} ${sanitisedPath(page)}.txt`;

        if (!filename.includes('about:blank')) {
            fs.writeFileSync(filename, textContent, 'utf8');
            console.log(`Text content saved to ${filename}`);
        }
    } catch (error) {
        console.error(`Failed to save text content: ${error.message}`);
    }

    return textContent;
}
