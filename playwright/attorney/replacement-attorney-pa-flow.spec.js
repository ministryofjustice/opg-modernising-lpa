import {expect, test} from '@playwright/test';
import {randomShareCode, screenshot, TestEmail} from '../../../test/playwright/e2e.js'
import {extractTextFromMainAndSave} from "../../../test/playwright/textExtractor.js";

test('replacement attorney property and affairs', async ({page}) => {
    const shareCode = randomShareCode()

    await page.goto(`/fixtures/attorney?redirect=&lpa-type=property-and-affairs&lpa-language=en&progress=signedByCertificateProvider&withShareCode=${shareCode}&email=${TestEmail}&options=is-replacement`);

    await page.goto(`/attorney-start`);
    await expect(page.locator('#main-content')).toContainText('Agree to be an attorney');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('textbox', {name: 'Enter code'}).fill(shareCode);
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.getByLabel('Success').getByRole('paragraph')).toContainText('We have identified your replacement attorney access code');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('link', {name: 'Confirm your details'}).click();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('radio', {name: 'English'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('link', {name: 'Read the LPA'}).click();

    await expect(page.locator('h1')).toContainText('Read the LPA');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('link', {name: 'Sign the LPA'}).click();

    await expect(page.locator('h1')).toContainText('Your legal rights and responsibilities');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('What happens when you sign the LPA');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue to signing page'}).click();

    await expect(page.locator('h1')).toContainText('Sign as a replacement attorney on this LPA');
    await page.getByRole('checkbox', {name: 'I, Blake Buckley, confirm'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Submit signature'}).click();

    await expect(page.locator('h1')).toContainText('You’ve formally agreed to be a replacement attorney');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Go to your dashboard'}).click();
});
