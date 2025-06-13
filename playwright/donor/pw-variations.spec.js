import {expect, test} from '@playwright/test';
import {screenshot} from '../e2e.js'
import {extractTextFromMainAndSave} from "../textExtractor.js";

test('donor personal welfare full journey', async ({page}) => {
    await page.goto('/start');
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('textbox', {name: 'First names'}).fill('Sam');
    await page.getByRole('textbox', {name: 'Last name'}).fill('Smith');
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('textbox', {name: 'Day'}).fill('02');
    await page.getByRole('textbox', {name: 'Month'}).fill('01');
    await page.getByRole('textbox', {name: 'Year'}).fill('2000');
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('radio', {name: 'Yes'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('textbox', {name: 'Postcode'}).fill('B14 7ED');
    await page.getByRole('button', {name: 'Find address'}).click();
    await page.getByLabel('Select an address').selectOption('{"line1":"1 RICHMOND PLACE","line2":"","line3":"","town":"BIRMINGHAM","postcode":"B14 7ED","country":"GB"}');
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: 'Yes'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('group', {name: 'Which language would you prefer us to use when we contact you?'}).getByLabel('English').check();
    await page.getByRole('group', {name: 'In which language would you'}).getByLabel('English').check();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('link', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('Choose which type of LPA you want to make');
    await page.getByRole('radio', {name: 'Personal welfare LPA'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('link', {name: 'Choose your attorneys'}).click();

    await expect(page.locator('h1')).toContainText('Choosing your attorneys');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('textbox', {name: 'First names'}).fill('Jessie');
    await page.getByRole('textbox', {name: 'Last name'}).fill('Jones');
    await page.getByRole('textbox', {name: 'Day'}).fill('02');
    await page.getByRole('textbox', {name: 'Month'}).fill('01');
    await page.getByRole('textbox', {name: 'Year'}).fill('2000');
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('radio', {name: 'Enter a new address'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('textbox', {name: 'Postcode'}).fill('B14 7ED');
    await page.getByRole('button', {name: 'Find address'}).click();
    await page.getByLabel('Select an address').selectOption('{"line1":"2 RICHMOND PLACE","line2":"","line3":"","town":"BIRMINGHAM","postcode":"B14 7ED","country":"GB"}');
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('radio', {name: 'Yes'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('textbox', {name: 'First names'}).fill('Robin');
    await page.getByRole('textbox', {name: 'Last name'}).fill('Redcar');
    await page.getByRole('textbox', {name: 'Day'}).fill('02');
    await page.getByRole('textbox', {name: 'Month'}).fill('01');
    await page.getByRole('textbox', {name: 'Year'}).fill('2000');
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('radio', {name: 'Use an address you’ve already'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: '2 RICHMOND PLACE BIRMINGHAM'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('radio', {name: 'No'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('h1')).toContainText('How your attorneys should make decisions');
    await page.getByRole('radio', {name: 'Jointly and severally', exact: true}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('link', {name: 'Choose your replacement'}).click();
    await page.getByRole('radio', {name: 'Yes, I want replacement'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('textbox', {name: 'First names'}).fill('Blake');
    await page.getByRole('textbox', {name: 'Last name'}).fill('Buckley');
    await page.getByRole('textbox', {name: 'Day'}).fill('02');
    await page.getByRole('textbox', {name: 'Month'}).fill('01');
    await page.getByRole('textbox', {name: 'Year'}).fill('2000');
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('radio', {name: 'Use an address you’ve already'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByText('2 RICHMOND PLACEBIRMINGHAMB14').click();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('radio', {name: 'Yes'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('textbox', {name: 'First names'}).fill('Taylor');
    await page.getByRole('textbox', {name: 'Last name'}).fill('Thompson');
    await page.getByRole('textbox', {name: 'Day'}).fill('02');
    await page.getByRole('textbox', {name: 'Month'}).fill('01');
    await page.getByRole('textbox', {name: 'Year'}).fill('2000');
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('radio', {name: 'Enter a new address'}).click();
    await page.getByRole('radio', {name: 'Use an address you’ve already'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: '2 RICHMOND PLACE BIRMINGHAM'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('radio', {name: 'No'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: 'All together, as soon as one'}).check();
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('link', {name: 'Life-sustaining treatment'}).click();

    await expect(page.locator('h1')).toContainText('Life-sustaining treatment');
    await page.getByRole('radio', {name: 'Yes - I do give my attorneys'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('link', {name: 'Add restrictions to the LPA'}).click();

    await expect(page.locator('h1')).toContainText('Restrict the decisions your attorneys can make');
    await page.getByText('Show me some examples').click();
    await screenshot(page)
    await extractTextFromMainAndSave(page)

    await expect(page.locator('h1')).toContainText('Restrict the decisions your attorneys can make');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('link', {name: 'Choose your certificate'}).click();
});
