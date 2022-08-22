Feature('Start');

Scenario('has a title', ({ I }) => {
    I.amOnPage('/')
    I.seeElement(
        locate('h1').withText('Make a lasting power of attorney')
    )
    I.runAccessibilityChecks()
});

Scenario('has a start button', ({ I }) => {
    I.amOnPage('/')
    I.seeElement(
        locate('a').withText('Start')
    )
    I.runAccessibilityChecks()
})
