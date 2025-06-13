import {expect, test} from '@playwright/test';
import {screenshot} from '../../../test/playwright/e2e.js'
import {extractTextFromMainAndSave} from "../../../test/playwright/textExtractor.js";

test('donor chooses a professional certifciate provider', async ({page}) => {
    await page.goto('/fixtures');
    await page.getByRole('radio', {name: 'Add restrictions to the LPA'}).check();
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('link', {name: 'Go to task list'}).click();
    await page.getByRole('link', {name: 'Choose your certificate'}).click();
    await page.getByRole('link', {name: 'Continue'}).click();

    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('textbox', {name: 'First names'}).fill('Charlie');
    await page.getByRole('textbox', {name: 'Last name'}).fill('Cooper');
    await page.getByRole('textbox', {name: 'UK mobile number'}).fill('07700900000');
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('How do you know Charlie Cooper, your certificate provider?');
    await page.getByRole('radio', {name: 'Professionally'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('How would Charlie like us to contact them?');
    await page.getByRole('radio', {name: 'By email'}).check();
    await page.getByRole('textbox', {name: 'Certificate provider’s email'}).click();
    await page.getByRole('textbox', {name: 'Certificate provider’s email'}).fill('simulate-delivered@notifications.service.gov.uk');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('What is Charlie Cooper’s work postcode?');
    await page.getByRole('textbox', {name: 'Postcode'}).click();
    await page.getByRole('textbox', {name: 'Postcode'}).fill('B14 7ED');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Find address'}).click();

    await expect(page.locator('h1')).toContainText('Select Charlie Cooper’s work address');
    await page.getByLabel('Select an address').selectOption('{"line1":"5 RICHMOND PLACE","line2":"","line3":"","town":"BIRMINGHAM","postcode":"B14 7ED","country":"GB"}');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Charlie Cooper’s work address');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('You’ve added a certificate provider');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Return to task list'}).click();
});
