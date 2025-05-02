const { expect } = require('@playwright/test');
const AxeBuilder = require('@axe-core/playwright').default;

export const
    TestEmail = 'simulate-delivered@notifications.service.gov.uk',
    TestEmail2 = 'simulate-delivered-2@notifications.service.gov.uk',
    TestMobile = '07700900000'

export function randomShareCode() {
    const characters = 'abcdefghijklmnpqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789'
    let result = [];

    for (let i = 0; i < 12; i++) {
        result.push(characters.charAt(Math.floor(Math.random() * characters.length)));
    }

    return result.join('');
}

export async function checkA11y(page) {
    const accessibilityScanResults = await new AxeBuilder({ page })
        .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
        .exclude('.govuk-phase-banner')
        .analyze();

    expect(accessibilityScanResults.violations).toEqual([]);
}

export async function visitLpa(path, page) {
    console.log(page.url())
    await page.goto(page.url().split('/').slice(0, 5).join('/') + path);
}

function sanitisedPath(page) {
    const parts = page.url().split('/')
    return parts[parts.length - 1].replace(/\//g, ' ')
}

export async function screenshot(page) {
    await page.screenshot({ path: `playwright/screenshots/${Date.now()} ${sanitisedPath(page)}.png`, fullPage: true });
}

export async function extractTextFromMainAndSave(page) {
    // Get text from main element with basic context
    const textContent = await page.evaluate(() => {

        const main = document.querySelector('main');
        if (!main) return 'No main element found';

        function parseChildrenNodes(element) {
            const isSummaryList = element.classList.contains('govuk-summary-list__row')
            const isListItem = element.tagName.toLowerCase() === 'li'

            let prependChar = ''
            if (isSummaryList) {
                prependChar = '|'
            } else if (isListItem) {
                prependChar = '•'
            }

            let paragraphText = '';

            // Process each child node of the paragraph
            for (const childNode of element.childNodes) {
                if (childNode.nodeType === Node.ELEMENT_NODE && childNode.tagName.toLowerCase() === 'a') {
                    // Link node
                    paragraphText += `[LINK: ${childNode.textContent.trim()}](${childNode.href})`;
                } else {
                    const textContent = childNode.textContent.replace(/\s+/g, " ").trim();
                    paragraphText += textContent === "." ? `${textContent} ` : `${prependChar} ${textContent} ` ;
                }
            }

            return paragraphText.trim() + '\n\n';
        }

        function getElementText(element) {
            const tagName = element.tagName.toLowerCase();
            const type = element.getAttribute('type');
            const classList = element.classList
            let text = '';

            // Skip script and style elements
            if (tagName === 'script' ||
                tagName === 'style' ||
                ['govuk-visually-hidden', 'govuk-hint', 'govuk-summary-list__row', 'govuk-!-display-none', 'app-dialog', 'govuk-warning-text__icon'].some(
                    cls => element.classList.contains(cls))
            ) {
                return '';
            }

            if (tagName === 'p' || classList.contains('govuk-summary-list__row') || tagName === 'li') {
                return parseChildrenNodes(element)
            }

            // Handle different element types
            switch (true) {
                // Buttons
                case (classList.contains('govuk-button') ||
                    tagName === 'button' ||
                    (tagName === 'input' && type === 'button') ||
                    (tagName === 'a' && element.getAttribute('role') === 'button')):
                    text += '[BUTTON: ' + element.textContent.trim() + ']\n\n';
                    break;

                // Radio buttons
                case (tagName === 'input' && type === 'radio'):
                    const radioChecked = element.checked ? '(checked)' : '(unchecked)';
                    const radioLabel = document.querySelector(`label[for="${element.id}"]`);
                    const radioLabelText = radioLabel ? radioLabel.textContent.trim() : element.value;
                    text += '[RADIO ' + radioChecked + ': ' + radioLabelText + ']\n\n';
                    break;

                // Checkboxes
                case (tagName === 'input' && type === 'checkbox'):
                    const checkboxChecked = element.checked ? '(checked)' : '(unchecked)';
                    const checkboxLabel = document.querySelector(`label[for="${element.id}"]`);
                    const checkboxLabelText = checkboxLabel ? checkboxLabel.textContent.trim() : element.value;
                    text += '[CHECKBOX ' + checkboxChecked + ': ' + checkboxLabelText + ']\n\n';
                    break;

                // Select dropdowns
                case (tagName === 'select'):
                    const options = Array.from(element.options)
                        .map(option => {
                            return (option.selected ? '> ' : '  ') + option.textContent.trim();
                        })
                        .join('\n    ');
                    text += '[SELECT: ' + (element.name || '') + ']\n    ' + options + '\n\n';
                    break;

                // Text inputs
                case (tagName === 'input' && !['checkbox', 'select', 'radio', 'textarea', 'hidden'].includes(type)):
                    const inputValue = element.value ? `: ${element.value}` : ': <NO VALUE>';
                    const inputLabel = document.querySelector(`label[for="${element.id}"]`);
                    const inputLabelText = inputLabel ? inputLabel.textContent.trim() : element.value;
                    const inputHint = element.parentNode.querySelector('.govuk-hint');

                    text += `${inputLabelText}\n`;
                    if (inputHint) {
                        text += `${inputHint.textContent.trim()}\n`;
                    }
                    text += `[INPUT` + inputValue + ']\n\n';
                    break;

                // Textareas
                case (tagName === 'textarea'):
                    const textareaLabel = document.querySelector(`label[for="${element.id}"]`);
                    const textareaValue = element.value ? ': ' + element.value : '';
                    text += `${textareaLabel}\n`;
                    text += '[TEXTAREA' + textareaValue + ']\n\n';
                    break;

                // Headings
                case ['h1', 'h2', 'h3', 'h4', 'h5', 'h6'].includes(tagName):
                    text += '\n' + element.textContent.trim() + '\n\n';
                    break;

                // Error messages
                case element.classList.contains('govuk-error-message'):
                    text += '[ERROR: ' + element.textContent.trim() + ']\n';
                    break;

                // Warnings
                case element.classList.contains('govuk-warning-text'):
                    text += `${element.textContent.replace(/\s+/g, " ").trim()} \n\n`;
                    break;

                // Default for other elements with direct text but exclude labels, options, visually hidden text/elements
                // and hints to stop duplication
                default:
                    if (element.childNodes.length === 0 ||
                        (element.childNodes.length === 1 && element.childNodes[0].nodeType === Node.TEXT_NODE)) {
                        const trimmed = element.textContent.trim();
                        if (trimmed && !['label', 'option'].includes(tagName)) {
                            text += trimmed + '\n\n';
                        }
                    }
                    break;
            }

            // Process children recursively
            for (const child of element.children) {
                text += getElementText(child);
            }

            return text;
        }

        return `[PAGE TITLE: ${document.title}]\n` + getElementText(main);
    });

    // Import fs module to write to file
    const fs = require('fs');
    fs.writeFileSync(`playwright/screenshots/${Date.now()} ${sanitisedPath(page)}.txt`, textContent, 'utf8');

    return textContent;
}
