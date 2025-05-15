import {expect, test} from '@playwright/test';
import {extractTextFromMainAndSave} from './TextExtractor';

test.describe('extractTextFromMainAndSave', () => {

    test('should handle missing main element', async ({ page }) => {
        await page.setContent('<body><h1>Outside Main</h1></body>');
        const textContent = await extractTextFromMainAndSave(page);
        expect(textContent).toBe('No main element found');
    });

    test('should handle empty main content', async ({ page }) => {
        await page.setContent('<main></main>');
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);
        expect(textContent).toBe(`[PAGE TITLE: ${defaultTitle}]\n\n`);
    });

    test('should extract text from a paragraph', async ({ page }) => {
        await page.setContent('<main><p>Some plain text.</p></main>');
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);
        expect(textContent).toBe(`[PAGE TITLE: ${defaultTitle}]\n\nSome plain text.\n\n`);
    });

    test('should prepend "|" and separate elements in a complex summary list row (2 dd)', async ({ page }) => {
        await page.setContent(`
            <main>
                <div class="govuk-summary-list__row">
                    <dt>Row title <span>with context</span> and text after</dt>
                    <dd>
                        Action text
                        <a href="https://www.example.com/a-link">link</a>
                        and more text
                    </dd>
                </div>
            </main>
        `);
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);
        expect(textContent).toBe(`[PAGE TITLE: ${defaultTitle}]\n\n| Row title with context and text after | Action text [LINK: link](https://www.example.com/a-link) and more text |\n\n`);
    });

    test('should prepend "|" and separate elements in a complex summary list row (3 dd)', async ({ page }) => {
        await page.setContent(`
            <main>
                <div class="govuk-summary-list__row">
                    <dt>Row title <span>with context</span> and text after</dt>
                    <dd>Row value</dd>
                    <dd>
                        Action text
                        <a href="https://www.example.com/a-link">link</a>
                        and more text
                    </dd>
                </div>
            </main>
        `);
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);
        expect(textContent).toBe(`[PAGE TITLE: ${defaultTitle}]\n\n| Row title with context and text after | Row value | Action text [LINK: link](https://www.example.com/a-link) and more text |\n\n`);
    });

    test('should prepend "•" for list items', async ({ page }) => {
        await page.setContent('<main><ul><li>List item.</li></ul></main>');
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);
        expect(textContent).toBe(`[PAGE TITLE: ${defaultTitle}]\n\n• List item.\n\n`);
    });

    test('should not add more than one "•" for list items with nested content', async ({ page }) => {
        await page.setContent(`
            <main>
                <ul>
                    <li>
                        List item with a
                        <a href="https://www.example.com/a-link">link</a>
                        and more content
                    </li>
                </ul>
            </main>
        `);
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);
        expect(textContent).toBe(`[PAGE TITLE: ${defaultTitle}]\n\n• List item with a [LINK: link](https://www.example.com/a-link) and more content\n\n`);
    });

    test('should extract link with markdown formatting within a paragraph', async ({ page }) => {
        await page.setContent('<main><p>See <a href="https://example.com">Example Link</a>.</p></main>');
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);
        expect(textContent).toBe(`[PAGE TITLE: ${defaultTitle}]\n\nSee [LINK: Example Link](https://example.com/) .\n\n`);
    });

    test('should handle mixed content in a paragraph', async ({ page }) => {
        await page.setContent(`
            <main>
                <p>
                    First part.
                    <a href="https://test.com">Test Link</a>
                          Second part.
                </p>
            </main>
        `);
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);
        expect(textContent).toBe(`[PAGE TITLE: ${defaultTitle}]\n\nFirst part. [LINK: Test Link](https://test.com/) Second part.\n\n`);
    });

    test('should handle multiple child nodes in a paragraph', async ({ page }) => {
        await page.setContent(`
            <main>
                <p>
                    <span>Span one.</span>
                    <span>Span two.</span>
                    <span>Span three.</span>
                </p>
            </main>
        `);
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);

        expect(textContent).toBe(`[PAGE TITLE: ${defaultTitle}]\n\nSpan one. Span two. Span three.\n\n`);
    });

    test('ignores certain elements', async ({ page }) => {
        await page.setContent(`
            <main>
                <p>Visible text before.</p>
                <label>Ignore this label</label>
                <option value="1">Ignore this option</option>
                <script>alert('ignore me')</script>
                <style>body { color: red; }</style>
                <p>Visible text after.</p>
            </main>
        `);
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);

        expect(textContent).toContain('Visible text before.\n\n');
        expect(textContent).toContain('Visible text after.\n\n');

        expect(textContent).not.toContain('Ignore this label');
        expect(textContent).not.toContain('Ignore this option');
        expect(textContent).not.toContain('ignore me');
        expect(textContent).not.toContain('body { color: red; }');

        const expectedText = `[PAGE TITLE: ${defaultTitle}]\n\nVisible text before.\n\nVisible text after.\n\n`;
        expect(textContent).toBe(expectedText);
    });

    test('ignores elements with certain classes', async ({ page }) => {
        await page.setContent(`
            <main>
                <p>Visible text before.</p>
                <div class="govuk-visually-hidden">Ignore visually hidden</div>
                <div class="govuk-hint">Ignore hint</div>
                <div class="govuk-!-display-none">Ignore display none</div>
                <div class="app-dialog">Ignore dialog</div>
                <div class="govuk-warning-text__icon">Ignore warning icon</div>
                <p>Visible text after.</p>
            </main>
        `);
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);

        expect(textContent).toContain('Visible text before.\n\n');
        expect(textContent).toContain('Visible text after.\n\n');

        expect(textContent).not.toContain('Ignore visually hidden');
        expect(textContent).not.toContain('Ignore hint');
        expect(textContent).not.toContain('Ignore display none');
        expect(textContent).not.toContain('Ignore dialog');
        expect(textContent).not.toContain('Ignore warning icon');

        const expectedText = `[PAGE TITLE: ${defaultTitle}]\n\nVisible text before.\n\nVisible text after.\n\n`;
        expect(textContent).toBe(expectedText);
    });


    test('handles buttons and button-like elements', async ({ page }) => {
        await page.setContent('<main><button>Click Me</button><input type="button" value="Submit"><a href="#" role="button">Go</a></main>');
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);

        expect(textContent).toContain('[BUTTON: Click Me]\n\n');
        expect(textContent).toContain('[BUTTON: Submit]\n\n');
        expect(textContent).toContain('[BUTTON: Go]\n\n');

        const expectedText = `[PAGE TITLE: ${defaultTitle}]\n\n[BUTTON: Click Me]\n\n[BUTTON: Submit]\n\n[BUTTON: Go]\n\n`;
        expect(textContent).toBe(expectedText);
    });

    test('handles radio and checkbox elements', async ({ page }) => {
        await page.setContent(`
            <main>
                <div>
                    <input type="radio" id="radio1" checked>
                    <label for="radio1">Option 1</label>
                </div>
                 <div>
                    <input type="checkbox" id="checkbox1">
                    <label for="checkbox1">Option 2</label>
                </div>
            </main>
        `);
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);

        expect(textContent).toContain("[RADIO (checked): Option 1]\n\n");
        expect(textContent).toContain("[CHECKBOX (unchecked): Option 2]\n\n");

        const expectedText = `[PAGE TITLE: ${defaultTitle}]\n\n[RADIO (checked): Option 1]\n\n[CHECKBOX (unchecked): Option 2]\n\n`;
        expect(textContent).toBe(expectedText);
    });

    test('handles selects', async ({ page }) => {
        await page.setContent(`
            <main>
                <select name="mySelect">
                    <option value="a">Option A</option>
                    <option value="b" selected>Option B</option>
                    <option value="c">Option C</option>
                </select>
            </main>
        `);
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);

        expect(textContent).toContain("[SELECT: mySelect]\nOption A\n> Option B\nOption C\n\n");

        const expectedText = `[PAGE TITLE: ${defaultTitle}]\n\n[SELECT: mySelect]\nOption A\n> Option B\nOption C\n\n`;
        expect(textContent).toBe(expectedText);
    });

    test('handles text inputs with label and hint', async ({ page }) => {
        await page.setContent(`
            <main>
                <div class="govuk-form-group">
                    <label class="govuk-label" for="myInput">Enter Name</label>
                    <div class="govuk-hint">Your full name</div>
                    <input class="govuk-input" id="myInput" type="text" value="John Doe">
                </div>
            </main>
        `);
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);

        expect(textContent).toContain("Enter Name\nYour full name\n[INPUT: John Doe]\n\n");

        const expectedText = `[PAGE TITLE: ${defaultTitle}]\n\nEnter Name\nYour full name\n[INPUT: John Doe]\n\n`;
        expect(textContent).toBe(expectedText);
    });

    test('handles textareas with label and hint', async ({ page }) => {
        await page.setContent(`
            <main>
                <div class="govuk-form-group">
                    <label class="govuk-label" for="myTextarea">Your Message</label>
                    <div class="govuk-hint">Keep it concise</div>
                    <textarea class="govuk-textarea" id="myTextarea">Hello World</textarea>
                </div>
            </main>
        `);
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);

        expect(textContent).toContain("Your Message\nKeep it concise\n[TEXTAREA: Hello World]\n\n");

        const expectedText = `[PAGE TITLE: ${defaultTitle}]\n\nYour Message\nKeep it concise\n[TEXTAREA: Hello World]\n\n`;
        expect(textContent).toBe(expectedText);
    });


    test('ignores hidden inputs', async ({ page }) => {
        await page.setContent('<main><p>Visible content.</p><input type="hidden" value="A value"></main>');
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);

        expect(textContent).toContain('Visible content.\n\n');
        expect(textContent).not.toContain('A value');

        const expectedText = `[PAGE TITLE: ${defaultTitle}]\n\nVisible content.\n\n`;
        expect(textContent).toBe(expectedText);
    });

    test('handles headers 1-6', async ({ page }) => {
        await page.setContent(`
            <main>
                <h1>  Header 1 content  </h1>
                <h2>Header 2</h2>
                <h3>Header 3 <span class="govuk-summary-card__title">with span content </span></h3>
                <h4>Header 4</h4>
                <h5>Header 5</h5>
                <h6>Header 6</h6>
            </main>
        `);
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);

        expect(textContent).toContain('Header 1 content\n\n');
        expect(textContent).toContain('Header 2\n\n');
        expect(textContent).toContain('Header 3 with span content\n\n');
        expect(textContent).toContain('Header 4\n\n');
        expect(textContent).toContain('Header 5\n\n');
        expect(textContent).toContain('Header 6\n\n');

        const expectedText = `[PAGE TITLE: ${defaultTitle}]\n\nHeader 1 content\n\nHeader 2\n\nHeader 3 with span content\n\nHeader 4\n\nHeader 5\n\nHeader 6\n\n`;
        expect(textContent).toBe(expectedText);
    });

    test('handles error elements', async ({ page }) => {
        await page.setContent('<main><div class="govuk-error-message">  Error message </div></main>');
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);

        expect(textContent).toContain('[ERROR: Error message]\n\n');

        const expectedText = `[PAGE TITLE: ${defaultTitle}]\n\n[ERROR: Error message]\n\n`;
        expect(textContent).toBe(expectedText);
    });

    test('handles warning elements', async ({ page }) => {
        await page.setContent('<main><div class="govuk-warning-text">    <strong class="govuk-warning-text__text">   Content  </strong></div></main>');
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);

        expect(textContent).toContain('[WARNING: Content]\n\n');

        const expectedText = `[PAGE TITLE: ${defaultTitle}]\n\n[WARNING: Content]\n\n`;
        expect(textContent).toBe(expectedText);
    });

    test('handles other elements with direct text nodes', async ({ page }) => {
        await page.setContent('<main><div>  Some text </div></main>');
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);

        expect(textContent).toContain(`[PAGE TITLE: ${defaultTitle}]`);
        expect(textContent).toContain('Some text\n\n');

        const expectedText = `[PAGE TITLE: ${defaultTitle}]\n\nSome text\n\n`;
        expect(textContent).toBe(expectedText);
    });

    test('handles other elements with child text nodes', async ({ page }) => {
        await page.setContent('<main><div><span>  Some text </span></div></main>');
        const defaultTitle = await page.title();
        const textContent = await extractTextFromMainAndSave(page);

        expect(textContent).toContain(`[PAGE TITLE: ${defaultTitle}]`);
        expect(textContent).toContain('Some text\n\n');

        const expectedText = `[PAGE TITLE: ${defaultTitle}]\n\nSome text\n\n`;
        expect(textContent).toBe(expectedText);
    });
});
