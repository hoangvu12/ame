function createSource({ fetch, clientFetch }) {
  var API = 'https://api.divineskins.gg/api';
  var IMG = 'https://images.divine-cdn.com';

  var CATEGORIES = [
    { label: 'All', value: '' },
    { label: 'Champion Mod', value: '1' },
    { label: 'Map Mod', value: '3' },
    { label: 'Sound Effects', value: '4' },
    { label: 'Font', value: '5' },
    { label: 'Announcer', value: '6' },
    { label: 'HUD/UI', value: '7' },
    { label: 'Other', value: '8' },
    { label: 'Recalls', value: '9' },
    { label: 'Chroma', value: '10' },
  ];

  var THEMES = [
    { label: 'All', value: '' },
    { label: 'Anime', value: '1' },
    { label: 'Game', value: '2' },
    { label: 'Edgy', value: '3' },
    { label: 'NSFW', value: '4' },
    { label: 'Meme', value: '5' },
    { label: 'Riot Style', value: '6' },
    { label: 'Chibi', value: '7' },
    { label: 'Other', value: '8' },
  ];

  var FEATURES = [
    { label: 'Sound Effects', value: '1' },
    { label: 'Visual Effects', value: '2' },
    { label: 'Animations', value: '3' },
    { label: 'Model', value: '4' },
    { label: 'Voice Over', value: '5' },
  ];

  var SORT_OPTIONS = [
    { label: 'Newest', value: 'approvedDate:desc' },
    { label: 'Oldest', value: 'approvedDate:asc' },
    { label: 'Most Downloaded', value: 'downloadCount:desc' },
    { label: 'Most Viewed', value: 'viewCount:desc' },
    { label: 'Most Liked', value: 'likeCount:desc' },
    { label: 'Recently Updated', value: 'lastUpdatedDate:desc' },
  ];

  function imgUrl(path) {
    if (!path) return '';
    return IMG + '/' + path;
  }

  function formatSize(bytes) {
    if (bytes >= 1048576) return (bytes / 1048576).toFixed(1) + ' MB';
    if (bytes >= 1024) return (bytes / 1024).toFixed(0) + ' KB';
    return bytes + ' B';
  }

  function mapItem(item) {
    return {
      id: item.slug,
      title: item.name,
      author: item.artistUsername || '',
      thumbnailUrl: imgUrl(item.imagePath),
      championId: 0,
      _artistUsername: item.artistUsername,
      _skinId: item.id,
    };
  }

  async function fetchPage(query, page, sortBy, direction, filters) {
    var params = 'page=' + page + '&size=20&sortBy=' + sortBy + '&direction=' + direction;
    if (filters.categoryId) params += '&categoryId=' + filters.categoryId;
    if (filters.themeId) params += '&themeId=' + filters.themeId;
    if (filters.featureIds) params += '&featureIds=' + encodeURIComponent(filters.featureIds);
    if (filters.championName) params += '&championName=' + encodeURIComponent(filters.championName);

    var path = '/catalog/skins';
    if (query) path += '/search/' + encodeURIComponent(query);

    var res = await fetch(API + path + '?' + params);
    var data = await res.json();
    var items = (data.content || []).map(mapItem);
    return { items: items, hasNextPage: !data.last };
  }

  async function getDownloadUrl(skinId, versionId) {
    var res = await fetch(API + '/celestial/manual/' + skinId + '/versions/' + versionId + '/download-url');
    var data = await res.json();
    return data.url || '';
  }

  return {
    id: 'divineskins',
    name: 'Divine Skins',
    baseUrl: 'https://divineskins.gg',
    lang: 'en',
    version: 1,
    iconUrl: 'https://divineskins.gg/celestial-logo.png',

    async getPopular(page) {
      return fetchPage('', page - 1, 'downloadCount', 'desc', {});
    },

    async getLatest(page) {
      return fetchPage('', page - 1, 'approvedDate', 'desc', {});
    },

    async search(query, page, filters) {
      var sort = (filters.sort || 'approvedDate:desc').split(':');
      var sortBy = sort[0];
      var direction = sort[1] || 'desc';
      var f = {
        categoryId: filters.categoryId || '',
        themeId: filters.themeId || '',
        featureIds: filters.featureIds || '',
        championName: filters.championName || '',
      };
      return fetchPage(query || '', page - 1, sortBy, direction, f);
    },

    async getDetails(skinMod) {
      var artist = skinMod._artistUsername || skinMod.author;
      var slug = skinMod.id;
      var res = await fetch(API + '/skins/by-slug/' + encodeURIComponent(artist) + '/' + encodeURIComponent(slug));
      var d = await res.json();

      var images = (d.galleryImages || []).map(function (p) { return imgUrl(p); });
      if (images.length === 0 && d.imagePath) {
        images = [imgUrl(d.imagePath)];
      }

      // Resolve real download URLs for each version
      var versions = d.versions || [];
      var downloads = [];
      for (var i = 0; i < versions.length; i++) {
        var v = versions[i];
        var url = await getDownloadUrl(d.id, v.id);
        if (url) {
          downloads.push({
            url: url,
            label: v.title + (v.fileSize ? ' (' + formatSize(v.fileSize) + ')' : ''),
            size: v.fileSize || null,
          });
        }
      }

      var desc = '';
      var stats = [];
      if (d.downloadCount) stats.push(d.downloadCount + ' downloads');
      if (d.viewCount) stats.push(d.viewCount + ' views');
      if (d.likeCount) stats.push(d.likeCount + ' likes');
      if (stats.length) desc += stats.join(' \u00B7 ') + '\n\n';
      if (d.description) desc += d.description;
      if (d.features && d.features.length) desc += '\n\nFeatures: ' + d.features.join(', ');
      if (d.themes && d.themes.length) desc += '\nThemes: ' + d.themes.join(', ');
      if (d.license) desc += '\nLicense: ' + d.license;

      return {
        id: skinMod.id,
        title: d.name,
        author: d.artistUsername || '',
        description: desc.trim(),
        championId: 0,
        images: images,
        downloads: downloads,
      };
    },

    getFilters: function () {
      return [
        { type: 'champion', name: 'Champion', key: 'champion', default: 0 },
        { type: 'sort', name: 'Sort by', key: 'sort', default: 'approvedDate:desc', options: SORT_OPTIONS },
        { type: 'select', name: 'Category', key: 'categoryId', default: '', options: CATEGORIES },
        { type: 'select', name: 'Theme', key: 'themeId', default: '', options: THEMES },
        { type: 'multiselect', name: 'Features', key: 'featureIds', default: '', options: FEATURES },
      ];
    },
  };
}
