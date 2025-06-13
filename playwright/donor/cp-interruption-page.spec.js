import {expect, test} from '@playwright/test';
import {screenshot} from '../../../test/playwright/e2e.js'
import {extractTextFromMainAndSave} from "../../../test/playwright/textExtractor.js";

test('donor sees certificate provider warning interruption page', async ({page}) => {
    await page.goto('/fixtures');
    await page.getByRole('radio', {name: 'Add a correspondent'}).check();
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('link', {name: 'Go to task list'}).click();
    await page.getByRole('link', {name: 'Choose your certificate'}).click();
    await page.getByRole('link', {name: 'Change   name for Charlie'}).click();
    await page.getByRole('textbox', {name: 'First names'}).fill('Sam');
    await page.getByRole('textbox', {name: 'Last name'}).fill('Smith');

    await expect(page.locator('h1')).toContainText('Your certificate provider’s details');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('You’ve added a certificate provider');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Return to task list'}).click();
    await page.getByRole('link', {name: 'Check and send to your'}).click();

    await expect(page.locator('#main-content')).toContainText('Confirm your certificate provider is not related to you or your attorneys');
    await page.getByRole('checkbox', {name: 'I confirm that my certificate'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();
});
