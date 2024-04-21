import React from 'react';

const ControllersTab = () => {
    return (
        <div id="Controllers" className="tabcontent">
            <button id="map">Map</button>
            <button id="defaultmap">DefaultMap</button>
            <button id="init">Init</button>
            <button id="pause">Pause</button>
            <button id="addGamepad">AddController</button>
            <table>
                <thead>
                    <tr>
                        <th>Controller 0</th>
                        <th>Controller 1</th>
                        <th>Controller 2</th>
                        <th>Controller 3</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>
                            <object id="svg-object0" type="image/svg+xml"></object>
                        </td>
                        <td>
                            <object id="svg-object1" type="image/svg+xml"></object>
                        </td>
                        <td>
                            <object id="svg-object2" type="image/svg+xml"></object>
                        </td>
                        <td>
                            <object id="svg-object3" type="image/svg+xml"></object>
                        </td>
                    </tr>
                </tbody>
            </table>
            <div className="table-container">
                <table id="defaultcontroller">
                    <caption>Default</caption>
                    <thead>
                        <tr>
                            <th>Button</th>
                            <th>Value</th>
                        </tr>
                    </thead>
                </table>
                {[0, 1, 2, 3].map((index) => (
                    <table key={`controller${index}`}>
                        <caption>Controller {index + 1}</caption>
                        <thead>
                            <tr>
                                <th>Button</th>
                                <th>Value</th>
                            </tr>
                        </thead>
                    </table>
                ))}
            </div>
        </div>
    );
};

export default ControllersTab;
