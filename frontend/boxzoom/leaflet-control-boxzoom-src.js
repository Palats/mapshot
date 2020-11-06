/*
 * L.Control.BoxZoom
 * A visible, clickable control for doing a box zoom.
 * https://github.com/gregallensworth/L.Control.BoxZoom
 */
L.Control.BoxZoom = L.Control.extend({
    options: {
        position: 'topright',
        title: 'Click here then draw a square on the map, to zoom in to an area',
        aspectRatio: null,
        divClasses: '',
        enableShiftDrag: false,
        iconClasses: '',
        keepOn: false,
    },
    initialize: function (options) {
        L.setOptions(this, options);
        this.map = null;
        this.active = false;
    },
    onAdd: function (map) {
        // add a linkage to the map, since we'll be managing map layers
        this.map = map;
        this.active = false;

        // create our button: uses FontAwesome cuz that font is... awesome
        // assign this here control as a property of the visible DIV, so we can be more terse when writing click-handlers on that visible DIV
        this.controlDiv = L.DomUtil.create('div', 'leaflet-control-boxzoom');

        // if we're not using an icon, add the background image class
        if (!this.options.iconClasses) {
            L.DomUtil.addClass(this.controlDiv, 'with-background-image');
        }
        if (this.options.divClasses) {
            L.DomUtil.addClass(this.controlDiv, this.options.divClasses);
        }
        this.controlDiv.control = this;
        this.controlDiv.title = this.options.title;
        this.controlDiv.innerHTML = ' ';
        L.DomEvent
            .addListener(this.controlDiv, 'mousedown', L.DomEvent.stopPropagation)
            .addListener(this.controlDiv, 'click', L.DomEvent.stopPropagation)
            .addListener(this.controlDiv, 'click', L.DomEvent.preventDefault)
            .addListener(this.controlDiv, 'click', function () {
                this.control.toggleState();
            });

        // start by toggling our state to off; this disables the boxZoom hooks on the map, in favor of this one
        this.setStateOff();

        if (this.options.iconClasses) {
            var iconElement = L.DomUtil.create('i', this.options.iconClasses, this.controlDiv);
            if (iconElement) {
                iconElement.style.color = this.options.iconColor || 'black';
                iconElement.style.textAlign = 'center';
                iconElement.style.verticalAlign = 'middle';
            } else {
                console.log('Unable to create element for icon');
            }
        }

        // if we're enforcing an aspect ratio, then monkey-patch the map's real BoxZoom control to support that
        // after all, this control is just a wrapper over the map's own BoxZoom behavior
        if (this.options.aspectRatio) {
            this.map.boxZoom.aspectRatio = this.options.aspectRatio;
            this.map.boxZoom._onMouseMove = this._boxZoomControlOverride_onMouseMove;
            this.map.boxZoom._onMouseUp = this._boxZoomControlOverride_onMouseUp;
        }

        // done!
        return this.controlDiv;
    },

    onRemove: function (map) {
        // on remove: if we had to monkey-patch the aspect-ratio stuff, undo that now
        if (this.options.aspectRatio) {
            delete this.map.boxZoom.aspectRatio;
            this.map.boxZoom._onMouseMove = L.Map.BoxZoom.prototype._onMouseMove;
            this.map.boxZoom._onMouseUp = L.Map.BoxZoom.prototype._onMouseUp;
        }
    },

    toggleState: function () {
        this.active ? this.setStateOff() : this.setStateOn();
    },
    setStateOn: function () {
        L.DomUtil.addClass(this.controlDiv, 'leaflet-control-boxzoom-active');
        this.active = true;
        this.map.dragging.disable();
        if (!this.options.enableShiftDrag) {
            this.map.boxZoom.addHooks();
        }

        this.map.on('mousedown', this.handleMouseDown, this);
        if (!this.options.keepOn) {
            this.map.on('boxzoomend', this.setStateOff, this);
        }

        L.DomUtil.addClass(this.map._container, 'leaflet-control-boxzoom-active');
    },
    setStateOff: function () {
        L.DomUtil.removeClass(this.controlDiv, 'leaflet-control-boxzoom-active');
        this.active = false;
        this.map.off('mousedown', this.handleMouseDown, this);
        this.map.dragging.enable();
        if (!this.options.enableShiftDrag) {
            this.map.boxZoom.removeHooks();
        }

        L.DomUtil.removeClass(this.map._container, 'leaflet-control-boxzoom-active');
    },

    handleMouseDown: function (event) {
        this.map.boxZoom._onMouseDown.call(this.map.boxZoom, { clientX: event.originalEvent.clientX, clientY: event.originalEvent.clientY, which: 1, shiftKey: true });
    },

    // monkey-patched applied to L.Map.BoxZoom to handle aspectRatio and to zoom to the drawn box instead of the mouseEvent point
    // in these methods,  "this" is not the control, but the map's boxZoom instance
    _boxZoomControlOverride_onMouseMove: function (e) {
        if (!this._moved) {
            this._box = L.DomUtil.create('div', 'leaflet-zoom-box', this._pane);
            L.DomUtil.setPosition(this._box, this._startLayerPoint);

            //TODO refactor: move cursor to styles
            this._container.style.cursor = 'crosshair';
            this._map.fire('boxzoomstart');
        }

        var startPoint = this._startLayerPoint,
            box = this._box,

            layerPoint = this._map.mouseEventToLayerPoint(e),
            offset = layerPoint.subtract(startPoint),

            newPos = new L.Point(
                Math.min(layerPoint.x, startPoint.x),
                Math.min(layerPoint.y, startPoint.y));

        L.DomUtil.setPosition(box, newPos);

        this._moved = true;

        var width = (Math.max(0, Math.abs(offset.x) - 4));  // from L.Map.BoxZoom, TODO refactor: remove hardcoded 4 pixels
        var height = (Math.max(0, Math.abs(offset.y) - 4));  // from L.Map.BoxZoom, TODO refactor: remove hardcoded 4 pixels

        if (this.aspectRatio) {
            height = width / this.aspectRatio;
        }

        box.style.width = width + 'px';
        box.style.height = height + 'px';
    },
    _boxZoomControlOverride_onMouseUp: function (e) {
        // the stock behavior is to generate a bbox based on the _startLayerPoint and the mouseUp event point
        // we don't want that; we specifically want to use the drawn box with the fixed aspect ratio

        // fetch the box and convert to a map bbox, before we clear it
        var ul = this._box._leaflet_pos;
        var lr = new L.Point(this._box._leaflet_pos.x + this._box.offsetWidth, this._box._leaflet_pos.y + this._box.offsetHeight);
        var nw = this._map.layerPointToLatLng(ul);
        var se = this._map.layerPointToLatLng(lr);
        if (nw.equals(se)) { return; }

        this._finish();

        var bounds = new L.LatLngBounds(nw, se);
        this._map.fitBounds(bounds);

        this._map.fire('boxzoomend', {
            boxZoomBounds: bounds
        });
    },
});
L.Control.boxzoom = function (options) {
    return new L.Control.BoxZoom(options);
}
