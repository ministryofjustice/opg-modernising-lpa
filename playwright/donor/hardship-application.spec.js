import {expect, test} from '@playwright/test';
import {screenshot} from '../../../test/playwright/e2e.js'
import {extractTextFromMainAndSave} from "../../../test/playwright/textExtractor.js";


test('donor applies for a hardship fee exemption', async ({page}) => {
    await page.goto('/fixtures');
    await page.getByRole('radio', {name: 'Check and send to your'}).check();
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('link', {name: 'Go to task list'}).click();
    await page.getByRole('link', {name: 'Pay for the LPA'}).click();
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: 'Yes'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('radio', {name: 'Hardship (fee waiver)'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('h1')).toContainText('Evidence required for a hardship application');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('How would you like to send us your evidence?');
    await page.getByRole('radio', {name: 'Send it by post'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.getByLabel('Important').getByRole('paragraph')).toContainText('We’ll review the evidence you send about your LPA fee');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Return to task list'}).click();
});
