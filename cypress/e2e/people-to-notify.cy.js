describe('People to notify', () => {
    context('something', () => {
        let person1
        let person2

        beforeEach(() => {
            cy.fixture('peopleToNotify.json').then(p => {
                person1 = p.person1
                person2 = p.person2
            })
        })

        it('can add people to notify', () => {
            cy.visit('/testing-start?redirect=/want-to-notify-people&cookies=accept');
            cy.injectAxe();

            cy.checkA11y(null, { rules: { region: { enabled: false } } });

            cy.get('input[name="want-to-notify"]').check('yes')
            cy.contains('button', 'Continue').click();

            cy.addPersonToNotify(person1)

            cy.injectAxe();
            cy.checkA11y(null, { rules: { region: { enabled: false } } });

            cy.contains('You have added 1 person to notify');

            cy.contains(`${person1.firstNames} ${person1.lastName}`);
            cy.contains(person1.address.line1);
            cy.contains(person1.address.town);
            cy.contains(person1.address.postcode);

            cy.get('input[name="add-person-to-notify"]').check('yes')
            cy.contains('button', 'Continue').click();

            cy.addPersonToNotify(person2)

            cy.contains('You have added 2 people to notify');

            cy.contains(`${person2.firstNames} ${person2.lastName}`);
            cy.contains(person2.address.line1);
            cy.contains(person2.address.town);
            cy.contains(person2.address.postcode);

            cy.get('input[name="add-person-to-notify"]').check('no')
            cy.contains('button', 'Continue').click();

            cy.visit('/task-list')

            cy.contains('a', 'People to notify').parent().parent().contains('Completed (2)')
        });
    })

});
